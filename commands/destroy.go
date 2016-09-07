package commands

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type Destroy struct {
	awsCredentialValidator awsCredentialValidator
	logger                 logger
	stdin                  io.Reader
	boshDeleter            boshDeleter
	vpcStatusChecker       vpcStatusChecker
	stackManager           stackManager
	stringGenerator        stringGenerator
	infrastructureManager  infrastructureManager
	keyPairDeleter         keyPairDeleter
	certificateDeleter     certificateDeleter
	stateStore             stateStore
}

type destroyConfig struct {
	NoConfirm bool
}

type keyPairDeleter interface {
	Delete(name string) error
}

type boshDeleter interface {
	Delete(boshInitManifest string, boshInitState boshinit.State, ec2PrivateKey string) error
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

func NewDestroy(awsCredentialValidator awsCredentialValidator, logger logger, stdin io.Reader,
	boshDeleter boshDeleter, vpcStatusChecker vpcStatusChecker, stackManager stackManager,
	stringGenerator stringGenerator, infrastructureManager infrastructureManager, keyPairDeleter keyPairDeleter,
	certificateDeleter certificateDeleter, stateStore stateStore) Destroy {
	return Destroy{
		awsCredentialValidator: awsCredentialValidator,
		logger:                 logger,
		stdin:                  stdin,
		boshDeleter:            boshDeleter,
		vpcStatusChecker:       vpcStatusChecker,
		stackManager:           stackManager,
		stringGenerator:        stringGenerator,
		infrastructureManager:  infrastructureManager,
		keyPairDeleter:         keyPairDeleter,
		certificateDeleter:     certificateDeleter,
		stateStore:             stateStore,
	}
}

func (d Destroy) Execute(subcommandFlags []string, state storage.State) error {
	err := d.awsCredentialValidator.Validate()
	if err != nil {
		return err
	}

	config, err := d.parseFlags(subcommandFlags)
	if err != nil {
		return err
	}

	if !config.NoConfirm {
		d.logger.Prompt("Are you sure you want to delete your infrastructure? This operation cannot be undone!")

		var proceed string
		fmt.Fscanln(d.stdin, &proceed)

		proceed = strings.ToLower(proceed)
		if proceed != "yes" && proceed != "y" {
			d.logger.Step("exiting")
			return nil
		}
	}

	stackExists := true
	stack, err := d.stackManager.Describe(state.Stack.Name)
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

	d.logger.Step("destroying BOSH director")
	state, err = d.deleteBOSH(stack, state)
	if err != nil {
		return err
	}

	if err := d.stateStore.Set(state); err != nil {
		return err
	}

	d.logger.Step("destroying AWS stack")
	state, err = d.deleteStack(stack, state)
	if err != nil {
		return err
	}

	if err := d.stateStore.Set(state); err != nil {
		return err
	}

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

	err = d.keyPairDeleter.Delete(state.KeyPair.Name)
	if err != nil {
		return err
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

	err := destroyFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (d Destroy) deleteBOSH(stack cloudformation.Stack, state storage.State) (storage.State, error) {
	emptyBOSH := storage.BOSH{}
	if reflect.DeepEqual(state.BOSH, emptyBOSH) {
		d.logger.Println("no BOSH director, skipping...")
		return state, nil
	}

	if err := d.boshDeleter.Delete(state.BOSH.Manifest, state.BOSH.State, state.KeyPair.PrivateKey); err != nil {
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
