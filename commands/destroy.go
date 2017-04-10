package commands

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

const (
	DestroyCommand = "destroy"
)

type Destroy struct {
	credentialValidator     credentialValidator
	logger                  logger
	stdin                   io.Reader
	boshManager             boshManager
	vpcStatusChecker        vpcStatusChecker
	stackManager            stackManager
	stringGenerator         stringGenerator
	infrastructureManager   infrastructureManager
	awsKeyPairDeleter       awsKeyPairDeleter
	gcpKeyPairDeleter       gcpKeyPairDeleter
	certificateDeleter      certificateDeleter
	stateStore              stateStore
	stateValidator          stateValidator
	terraformManager        terraformManager
	networkInstancesChecker networkInstancesChecker
}

type destroyConfig struct {
	NoConfirm     bool
	SkipIfMissing bool
}

type awsKeyPairDeleter interface {
	Delete(name string) error
}

type gcpKeyPairDeleter interface {
	Delete(publicKey string) error
}

type vpcStatusChecker interface {
	ValidateSafeToDelete(string) error
}

type stackManager interface {
	Describe(string) (cloudformation.Stack, error)
}

type stringGenerator interface {
	Generate(prefix string, length int) (string, error)
}

type certificateDeleter interface {
	Delete(certificateName string) error
}

type stateValidator interface {
	Validate() error
}

type networkInstancesChecker interface {
	ValidateSafeToDelete(networkName string) error
}

func NewDestroy(credentialValidator credentialValidator, logger logger, stdin io.Reader,
	boshManager boshManager, vpcStatusChecker vpcStatusChecker, stackManager stackManager,
	stringGenerator stringGenerator, infrastructureManager infrastructureManager, awsKeyPairDeleter awsKeyPairDeleter,
	gcpKeyPairDeleter gcpKeyPairDeleter, certificateDeleter certificateDeleter, stateStore stateStore, stateValidator stateValidator,
	terraformManager terraformManager, networkInstancesChecker networkInstancesChecker) Destroy {
	return Destroy{
		credentialValidator:     credentialValidator,
		logger:                  logger,
		stdin:                   stdin,
		boshManager:             boshManager,
		vpcStatusChecker:        vpcStatusChecker,
		stackManager:            stackManager,
		stringGenerator:         stringGenerator,
		infrastructureManager:   infrastructureManager,
		awsKeyPairDeleter:       awsKeyPairDeleter,
		gcpKeyPairDeleter:       gcpKeyPairDeleter,
		certificateDeleter:      certificateDeleter,
		stateStore:              stateStore,
		stateValidator:          stateValidator,
		terraformManager:        terraformManager,
		networkInstancesChecker: networkInstancesChecker,
	}
}

func (d Destroy) Execute(subcommandFlags []string, state storage.State) error {
	if !state.NoDirector {
		err := fastFailBOSHVersion(d.boshManager)
		if err != nil {
			return err
		}
	}

	if state.IAAS == "gcp" {
		err := d.terraformManager.ValidateVersion()
		if err != nil {
			return err
		}
	}

	config, err := d.parseFlags(subcommandFlags)
	if err != nil {
		return err
	}

	if config.SkipIfMissing && state.EnvID == "" {
		d.logger.Step("state file not found, and --skip-if-missing flag provided, exiting")
		return nil
	}

	err = d.stateValidator.Validate()
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "aws":
		err = d.credentialValidator.ValidateAWS()
		if err != nil {
			return err
		}
	case "gcp":
		err = d.credentialValidator.ValidateGCP()
		if err != nil {
			return err
		}
	}

	var terraformOutputs terraform.Outputs
	if state.IAAS == "gcp" {
		domainExists := false
		if state.LB.Domain != "" {
			domainExists = true
		}

		terraformOutputs, err = d.terraformManager.GetOutputs(state.TFState, state.LB.Type, domainExists)
		if err != nil {
			return err
		}

		err = d.networkInstancesChecker.ValidateSafeToDelete(terraformOutputs.NetworkName)
		if err != nil {
			return err
		}
	}

	if !config.NoConfirm {
		d.logger.Prompt(fmt.Sprintf("Are you sure you want to delete infrastructure for %q? This operation cannot be undone!", state.EnvID))

		var proceed string
		fmt.Fscanln(d.stdin, &proceed)

		proceed = strings.ToLower(proceed)
		if proceed != "yes" && proceed != "y" {
			d.logger.Step("exiting")
			return nil
		}
	}

	var stack cloudformation.Stack
	if state.IAAS == "aws" {
		stackExists := true
		var err error
		stack, err = d.stackManager.Describe(state.Stack.Name)
		switch err {
		case cloudformation.StackNotFound:
			stackExists = false
		case nil:
			break
		default:
			return err
		}

		if stackExists {
			var vpcID = stack.Outputs["VPCID"]
			if err := d.vpcStatusChecker.ValidateSafeToDelete(vpcID); err != nil {
				return err
			}
		}
	}

	state, err = d.deleteBOSH(state, stack, terraformOutputs)
	switch err.(type) {
	case bosh.ManagerDeleteError:
		mdErr := err.(bosh.ManagerDeleteError)
		setErr := d.stateStore.Set(mdErr.State())
		if setErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(setErr)
			return errorList
		}
		return err
	case error:
		return err
	}

	if err := d.stateStore.Set(state); err != nil {
		return err
	}

	if state.IAAS == "aws" {
		d.logger.Step("destroying AWS stack")
		state, err = d.deleteStack(stack, state)
		if err != nil {
			return err
		}
	}

	if state.IAAS == "gcp" {
		state, err = d.terraformManager.Destroy(state)
		if err != nil {
			return handleTerraformError(err, d.stateStore)
		}
	}

	if err := d.stateStore.Set(state); err != nil {
		return err
	}

	if state.IAAS == "aws" {
		if state.Stack.CertificateName != "" {
			d.logger.Step("deleting certificate")
			err = d.certificateDeleter.Delete(state.Stack.CertificateName)
			if err != nil {
				return err
			}

			state.Stack.CertificateName = ""

			if err := d.stateStore.Set(state); err != nil {
				return err
			}
		}
	}

	switch state.IAAS {
	case "aws":
		err = d.awsKeyPairDeleter.Delete(state.KeyPair.Name)
		if err != nil {
			return err
		}

	case "gcp":
		err = d.gcpKeyPairDeleter.Delete(state.KeyPair.PublicKey)
		if err != nil {
			return err
		}
	}

	err = d.stateStore.Set(storage.State{})
	if err != nil {
		return err
	}

	return nil
}

func (d Destroy) parseFlags(subcommandFlags []string) (destroyConfig, error) {
	destroyFlags := flags.New("destroy")

	config := destroyConfig{}
	destroyFlags.Bool(&config.NoConfirm, "n", "no-confirm", false)
	destroyFlags.Bool(&config.SkipIfMissing, "", "skip-if-missing", false)

	err := destroyFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (d Destroy) deleteBOSH(state storage.State, stack cloudformation.Stack, terraformOutputs terraform.Outputs) (storage.State, error) {
	emptyBOSH := storage.BOSH{}
	if reflect.DeepEqual(state.BOSH, emptyBOSH) {
		d.logger.Println("no BOSH director, skipping...")
		return state, nil
	}

	d.logger.Step("destroying bosh director")

	err := d.boshManager.Delete(state)
	if err != nil {
		return state, err
	}

	state.BOSH = storage.BOSH{}

	return state, nil
}

func (d Destroy) deleteStack(stack cloudformation.Stack, state storage.State) (storage.State, error) {
	if state.Stack.Name == "" {
		d.logger.Println("no AWS stack, skipping...")
		return state, nil
	}

	if err := d.infrastructureManager.Delete(state.Stack.Name); err != nil {
		return state, err
	}

	state.Stack.Name = ""
	state.Stack.LBType = ""

	return state, nil
}
