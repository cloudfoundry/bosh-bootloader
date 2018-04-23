package commands

import (
	"errors"
	"fmt"

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

	state, err := v.terraformManager.Validate(state)
	if err != nil {
		return handleTerraformError(err, state, v.stateStore)
	}

	state.NoDirector = false

	err = v.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after terraform validate: %s", err)
	}

	return nil
}
