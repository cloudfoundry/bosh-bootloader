package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPDeleteLBs struct {
	cloudConfigManager   cloudConfigManager
	stateStore           stateStore
	environmentValidator environmentValidator
	terraformManager     terraformApplier
}

func NewGCPDeleteLBs(stateStore stateStore, environmentValidator environmentValidator,
	terraformManager terraformApplier, cloudConfigManager cloudConfigManager) GCPDeleteLBs {
	return GCPDeleteLBs{
		stateStore:           stateStore,
		environmentValidator: environmentValidator,
		terraformManager:     terraformManager,
		cloudConfigManager:   cloudConfigManager,
	}
}

func (g GCPDeleteLBs) Execute(state storage.State) error {
	err := g.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	if err := g.environmentValidator.Validate(state); err != nil {
		return err
	}

	state.LB.Type = ""

	if !state.NoDirector {
		err = g.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	state, err = g.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, g.stateStore)
	}

	err = g.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
