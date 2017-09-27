package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AWSDeleteLBs struct {
	cloudConfigManager   cloudConfigManager
	stateStore           stateStore
	environmentValidator EnvironmentValidator
	terraformManager     terraformApplier
}

type deleteLBsConfig struct {
	skipIfMissing bool
}

func NewAWSDeleteLBs(cloudConfigManager cloudConfigManager, stateStore stateStore,
	environmentValidator EnvironmentValidator, terraformManager terraformApplier) AWSDeleteLBs {
	return AWSDeleteLBs{
		cloudConfigManager:   cloudConfigManager,
		stateStore:           stateStore,
		environmentValidator: environmentValidator,
		terraformManager:     terraformManager,
	}
}

func (c AWSDeleteLBs) Execute(state storage.State) error {
	err := c.environmentValidator.Validate(state)
	if err != nil {
		return err
	}

	if !lbExists(state.LB.Type) {
		return LBNotFound
	}

	state.LB = storage.LB{}

	if !state.NoDirector {
		err = c.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	state, err = c.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, c.stateStore)
	}

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
