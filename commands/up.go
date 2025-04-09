package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Up struct {
	plan                 plan
	boshManager          boshManager
	cloudConfigManager   cloudConfigManager
	runtimeConfigManager runtimeConfigManager
	stateStore           stateStore
	terraformManager     terraformManager
}

func NewUp(plan plan, boshManager boshManager,
	cloudConfigManager cloudConfigManager,
	runtimeConfigManager runtimeConfigManager,
	stateStore stateStore, terraformManager terraformManager) Up {
	return Up{
		plan:                 plan,
		boshManager:          boshManager,
		cloudConfigManager:   cloudConfigManager,
		runtimeConfigManager: runtimeConfigManager,
		stateStore:           stateStore,
		terraformManager:     terraformManager,
	}
}

func (u Up) CheckFastFails(args []string, state storage.State) error {
	return u.plan.CheckFastFails(args, state)
}

func (u Up) Execute(args []string, state storage.State) error {
	config, err := u.ParseArgs(args, state)
	if err != nil {
		return err
	}

	if !u.plan.IsInitialized(state) {
		planState, err := u.plan.InitializePlan(config, state)
		if err != nil {
			return err
		}
		state = planState
	}

	state, err = u.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, state, u.stateStore)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after terraform apply: %s", err) //nolint:staticcheck
	}

	terraformOutputs, err := u.terraformManager.GetOutputs()
	if err != nil {
		return fmt.Errorf("Parse terraform outputs: %s", err) //nolint:staticcheck
	}

	state, err = u.boshManager.CreateJumpbox(state, terraformOutputs)
	switch err.(type) { //nolint:staticcheck
	case bosh.ManagerCreateError:
		bcErr := err.(bosh.ManagerCreateError) //nolint:staticcheck
		if setErr := u.stateStore.Set(bcErr.State()); setErr != nil {
			return fmt.Errorf("Save state after jumpbox create error: %s, %s", err, setErr) //nolint:staticcheck
		}
		return fmt.Errorf("Create jumpbox: %s", err) //nolint:staticcheck
	case error:
		return fmt.Errorf("Create jumpbox: %s", err) //nolint:staticcheck
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after create jumpbox: %s", err) //nolint:staticcheck
	}

	state, err = u.boshManager.CreateDirector(state, terraformOutputs)
	switch err.(type) { //nolint:staticcheck
	case bosh.ManagerCreateError:
		bcErr := err.(bosh.ManagerCreateError) //nolint:staticcheck
		if setErr := u.stateStore.Set(bcErr.State()); setErr != nil {
			return fmt.Errorf("Save state after bosh director create error: %s, %s", err, setErr) //nolint:staticcheck
		}
		return fmt.Errorf("Create bosh director: %s", err) //nolint:staticcheck
	case error:
		return fmt.Errorf("Create bosh director: %s", err) //nolint:staticcheck
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after create director: %s", err) //nolint:staticcheck
	}

	err = u.cloudConfigManager.Update(state)
	if err != nil {
		return fmt.Errorf("Update cloud config: %s", err) //nolint:staticcheck
	}

	err = u.runtimeConfigManager.Update(state)
	if err != nil {
		return fmt.Errorf("Update runtime config: %s", err) //nolint:staticcheck
	}

	return nil
}

func (u Up) ParseArgs(args []string, state storage.State) (PlanConfig, error) {
	return u.plan.ParseArgs(args, state)
}
