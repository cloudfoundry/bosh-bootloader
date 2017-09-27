package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Destroy struct {
	logger                   logger
	stdin                    io.Reader
	boshManager              boshManager
	stateStore               stateStore
	stateValidator           stateValidator
	terraformManager         terraformDestroyer
	networkDeletionValidator NetworkDeletionValidator
}

type destroyConfig struct {
	NoConfirm     bool
	SkipIfMissing bool
}

type NetworkDeletionValidator interface {
	ValidateSafeToDelete(networkName string, envID string) error
}

func NewDestroy(logger logger, stdin io.Reader,
	boshManager boshManager, stateStore stateStore, stateValidator stateValidator,
	terraformManager terraformDestroyer, networkDeletionValidator NetworkDeletionValidator) Destroy {
	return Destroy{
		logger:                   logger,
		stdin:                    stdin,
		boshManager:              boshManager,
		stateStore:               stateStore,
		stateValidator:           stateValidator,
		terraformManager:         terraformManager,
		networkDeletionValidator: networkDeletionValidator,
	}
}

func (d Destroy) CheckFastFails(subcommandFlags []string, state storage.State) error {
	if !state.NoDirector {
		err := fastFailBOSHVersion(d.boshManager)
		if err != nil {
			return err
		}
	}

	err := d.terraformManager.ValidateVersion()
	if err != nil {
		return err
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

	terraformOutputs, err := d.terraformManager.GetOutputs(state)
	if err != nil {
		return nil
	}

	var networkName string
	if state.IAAS == "gcp" {
		output, ok := terraformOutputs["network_name"]
		if !ok {
			return nil
		}
		networkName = output.(string)
	} else if state.IAAS == "aws" {
		output, ok := terraformOutputs["vpc_id"]
		if !ok {
			return nil
		}
		networkName = output.(string)
	} else if state.IAAS == "azure" {
		return nil
	}

	err = d.networkDeletionValidator.ValidateSafeToDelete(networkName, state.EnvID)
	if err != nil {
		return err
	}

	return nil
}

func (d Destroy) Execute(subcommandFlags []string, state storage.State) error {
	config, err := d.parseFlags(subcommandFlags)
	if err != nil {
		return err
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

	terraformOutputs, err := d.terraformManager.GetOutputs(state)
	if err != nil {
		return err
	}

	state, err = d.deleteBOSH(state, terraformOutputs)
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

	state, err = d.terraformManager.Destroy(state)
	if err != nil {
		return handleTerraformError(err, d.stateStore)
	}

	if err := d.stateStore.Set(state); err != nil {
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

func (d Destroy) deleteBOSH(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	if state.NoDirector {
		d.logger.Println("no BOSH director, skipping...")
		return state, nil
	}

	if !state.BOSH.IsEmpty() {
		d.logger.Step("destroying bosh director")

		err := d.boshManager.DeleteDirector(state, terraformOutputs)
		if err != nil {
			return state, err
		}

		state.BOSH = storage.BOSH{}
	}

	if !state.Jumpbox.IsEmpty() {
		err := d.boshManager.DeleteJumpbox(state, terraformOutputs)
		if err != nil {
			return state, err
		}

		state.Jumpbox = storage.Jumpbox{}
	}

	return state, nil
}
