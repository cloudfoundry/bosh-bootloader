package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AzureDeleteLBs struct {
	cloudConfigManager cloudConfigManager
	stateStore         stateStore
	terraformManager   terraformApplier
}

func NewAzureDeleteLBs(cloudConfigManager cloudConfigManager, stateStore stateStore, terraformManager terraformApplier) AzureDeleteLBs {
	return AzureDeleteLBs{
		cloudConfigManager: cloudConfigManager,
		stateStore:         stateStore,
		terraformManager:   terraformManager,
	}
}

func (a AzureDeleteLBs) Execute(state storage.State) error {
	err := a.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	state.LB = storage.LB{}

	if !state.NoDirector {
		err = a.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	state, err = a.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, a.stateStore)
	}

	err = a.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
