package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const (
	DeleteLBsCommand = "delete-lbs"
)

type AWSDeleteLBs struct {
	credentialValidator  credentialValidator
	logger               logger
	cloudConfigManager   cloudConfigManager
	stateStore           stateStore
	environmentValidator environmentValidator
	terraformManager     terraformApplier
}

type deleteLBsConfig struct {
	skipIfMissing bool
}

func NewAWSDeleteLBs(credentialValidator credentialValidator,
	logger logger,
	cloudConfigManager cloudConfigManager, stateStore stateStore,
	environmentValidator environmentValidator, terraformManager terraformApplier,
) AWSDeleteLBs {
	return AWSDeleteLBs{
		credentialValidator:  credentialValidator,
		logger:               logger,
		cloudConfigManager:   cloudConfigManager,
		stateStore:           stateStore,
		environmentValidator: environmentValidator,
		terraformManager:     terraformManager,
	}
}

func (c AWSDeleteLBs) Execute(state storage.State) error {
	err := c.credentialValidator.Validate()
	if err != nil {
		return err
	}

	err = c.environmentValidator.Validate(state)
	if err != nil {
		return err
	}

	if state.Stack.LBType != "" {
		state.LB.Type = state.Stack.LBType

		state, err = c.terraformManager.Apply(state)
		if err != nil {
			return handleTerraformError(err, c.stateStore)
		}
	}

	if !lbExists(state.LB.Type) {
		if !lbExists(state.Stack.LBType) {
			return LBNotFound
		}
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
