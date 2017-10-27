package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type Plan struct {
	up                 up
	boshManager        boshManager
	cloudConfigManager cloudConfigManager
	stateStore         stateStore
	envIDManager       envIDManager
	terraformManager   terraformManager
}

func NewPlan(up up, boshManager boshManager, cloudConfigManager cloudConfigManager,
	stateStore stateStore, envIDManager envIDManager, terraformManager terraformManager) Plan {
	return Plan{
		up:                 up,
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		stateStore:         stateStore,
		envIDManager:       envIDManager,
		terraformManager:   terraformManager,
	}
}

func (p Plan) CheckFastFails(args []string, state storage.State) error {
	return p.up.CheckFastFails(args, state)
}

func (p Plan) ParseArgs(args []string, state storage.State) (UpConfig, error) {
	return p.up.ParseArgs(args, state)
}

func (p Plan) Execute(args []string, state storage.State) error {
	config, err := p.ParseArgs(args, state)
	if err != nil {
		return err
	}

	if config.NoDirector {
		if !state.BOSH.IsEmpty() {
			return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}
		state.NoDirector = true
	}

	err = p.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state: %s", err)
	}

	if err := p.terraformManager.Init(state); err != nil {
		return fmt.Errorf("Terraform manager init: %s", err)
	}

	if state.NoDirector {
		return nil
	}

	if err := p.boshManager.InitializeJumpbox(state, terraform.Outputs{}); err != nil {
		return fmt.Errorf("Bosh manager initialize jumpbox: %s", err)
	}

	if err := p.boshManager.InitializeDirector(state, terraform.Outputs{}); err != nil {
		return fmt.Errorf("Bosh manager initialize director: %s", err)
	}

	return nil
}
