package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type Destroy struct {
	plan                     plan
	logger                   logger
	boshManager              boshManager
	stateStore               stateStore
	stateValidator           stateValidator
	terraformManager         terraformManager
	networkDeletionValidator NetworkDeletionValidator
}

type destroyConfig struct {
	NoConfirm bool
}

type NetworkDeletionValidator interface {
	ValidateSafeToDelete(networkName string, envID string) error
}

func NewDestroy(plan plan, logger logger, boshManager boshManager, stateStore stateStore,
	stateValidator stateValidator, terraformManager terraformManager,
	networkDeletionValidator NetworkDeletionValidator) Destroy {
	return Destroy{
		plan:                     plan,
		logger:                   logger,
		boshManager:              boshManager,
		stateStore:               stateStore,
		stateValidator:           stateValidator,
		terraformManager:         terraformManager,
		networkDeletionValidator: networkDeletionValidator,
	}
}

func (d Destroy) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := fastFailBOSHVersion(d.boshManager)
	if err != nil {
		return err
	}

	err = d.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	err = d.stateValidator.Validate()
	if _, ok := err.(NoBBLStateError); ok {
		d.logger.Println(err.Error())
		return ExitSuccessfully{}
	}

	if err != nil {
		return err
	}

	isPaved, _ := d.terraformManager.IsPaved()
	if !isPaved {
		return nil
	}

	terraformOutputs, err := d.terraformManager.GetOutputs()
	if err != nil {
		return nil
	}

	var networkName string
	switch state.IAAS {
	case "gcp":
		networkName = terraformOutputs.GetString("network_name")
	case "aws":
		networkName = terraformOutputs.GetString("vpc_id")
	case "azure":
		networkName = terraformOutputs.GetString("bosh_network_name")
	}

	if networkName == "" {
		return nil
	}

	err = d.networkDeletionValidator.ValidateSafeToDelete(networkName, state.EnvID)
	if err != nil {
		return err
	}

	return nil
}

func (d Destroy) Execute(subcommandFlags []string, state storage.State) error {
	proceed := d.logger.Prompt(fmt.Sprintf("Are you sure you want to delete infrastructure for %q? This operation cannot be undone!", state.EnvID))
	if !proceed {
		d.logger.Step("exiting")
		return nil
	}

	if !d.plan.IsInitialized(state) {
		planConfig := PlanConfig{
			Name: state.EnvID,
			LB:   state.LB,
		}

		var err error
		state, err = d.plan.InitializePlan(planConfig, state)
		if err != nil {
			return fmt.Errorf("Initialize plan during destroy: %s", err)
		}
	}

	isPaved, err := d.terraformManager.IsPaved()
	if err != nil {
		return err
	}

	if !isPaved {
		if err := d.stateStore.Set(storage.State{}); err != nil {
			return err
		}
		return nil
	}

	terraformOutputs, err := d.terraformManager.GetOutputs()
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

	if err = d.terraformManager.Setup(state); err != nil {
		return err
	}

	state, err = d.terraformManager.Destroy(state)
	if err != nil {
		return handleTerraformError(err, state, d.stateStore)
	}

	if err := d.stateStore.Set(storage.State{}); err != nil {
		return err
	}

	return nil
}

func (d Destroy) deleteBOSH(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	err := d.boshManager.CleanUpDirector(state)
	if err != nil {
		return state, err
	}

	err = d.boshManager.DeleteDirector(state, terraformOutputs)
	if err != nil {
		return state, err
	}

	state.BOSH = storage.BOSH{}

	err = d.boshManager.DeleteJumpbox(state, terraformOutputs)
	if err != nil {
		return state, err
	}

	state.Jumpbox = storage.Jumpbox{}

	return state, nil
}
