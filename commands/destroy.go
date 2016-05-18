package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type Destroy struct {
	logger                logger
	stdin                 io.Reader
	boshDeleter           boshDeleter
	vpcStatusChecker      vpcStatusChecker
	stackManager          stackManager
	stringGenerator       stringGenerator
	infrastructureManager infrastructureManager
	keyPairDeleter        keyPairDeleter
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

func NewDestroy(logger logger, stdin io.Reader, boshDeleter boshDeleter, vpcStatusChecker vpcStatusChecker, stackManager stackManager, stringGenerator stringGenerator, infrastructureManager infrastructureManager, keyPairDeleter keyPairDeleter) Destroy {
	return Destroy{
		logger:                logger,
		stdin:                 stdin,
		boshDeleter:           boshDeleter,
		vpcStatusChecker:      vpcStatusChecker,
		stackManager:          stackManager,
		stringGenerator:       stringGenerator,
		infrastructureManager: infrastructureManager,
		keyPairDeleter:        keyPairDeleter,
	}
}

func (d Destroy) Execute(globalFlags GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	config, err := d.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	if !config.NoConfirm {
		d.logger.Prompt("Are you sure you want to delete your infrastructure? This operation cannot be undone!")

		var proceed string
		fmt.Fscanln(d.stdin, &proceed)

		proceed = strings.ToLower(proceed)
		if proceed != "yes" && proceed != "y" {
			d.logger.Step("exiting")
			return state, nil
		}
	}

	d.logger.Step("destroying BOSH director and AWS stack")

	stack, err := d.stackManager.Describe(state.Stack.Name)
	if err != nil {
		return state, err
	}

	var vpcID = stack.Outputs["VPCID"]
	if err := d.vpcStatusChecker.ValidateSafeToDelete(vpcID); err != nil {
		return state, err
	}

	state, err = d.deleteBOSH(stack, state)
	if err != nil {
		return state, err
	}

	if err := d.infrastructureManager.Delete(state.Stack.Name); err != nil {
		return state, err
	}

	err = d.keyPairDeleter.Delete(state.KeyPair.Name)
	if err != nil {
		return state, err
	}

	return storage.State{}, nil
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
	if err := d.boshDeleter.Delete(state.BOSH.Manifest, state.BOSH.State, state.KeyPair.PrivateKey); err != nil {
		return state, err
	}

	return state, nil
}
