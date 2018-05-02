package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Validate struct {
	plan             plan
	stateStore       stateStore
	terraformManager terraformManager
}

func NewValidate(plan plan, stateStore stateStore, terraformManager terraformManager) Validate {
	return Validate{
		plan:             plan,
		stateStore:       stateStore,
		terraformManager: terraformManager,
	}
}

func (v Validate) CheckFastFails(args []string, state storage.State) error {
	if state.IAAS == "" {
		return errors.New("bbl state has not been initialized yet, please run bbl plan")
	}
	return v.plan.CheckFastFails(args, state)
}

func (v Validate) Execute(args []string, state storage.State) error {
	if !v.plan.IsInitialized(state) {
		return errors.New("bbl state has not been initialized yet, please run bbl plan")
	}

	err := v.terraformManager.Init(state)
	if err != nil {
		return handleTerraformError(err, state, v.stateStore)
	}

	state, err = v.terraformManager.Validate(state)
	if err != nil {
		return handleTerraformError(err, state, v.stateStore)
	}

	return nil
}
