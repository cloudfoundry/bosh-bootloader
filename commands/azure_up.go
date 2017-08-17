package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type azureClient interface {
	ValidateCredentials(subscriptionID, tenantID, clientID, clientSecret string) error
}

type AzureUpConfig struct {
	Name       string
	NoDirector bool
}

type AzureUp struct {
	azureClient      azureClient
	logger           logger
	envIDManager     envIDManager
	stateStore       stateStore
	terraformManager terraformApplier
}

func NewAzureUp(azureClient azureClient, logger logger, envIDManager envIDManager, stateStore stateStore, terraformManager terraformApplier) AzureUp {
	return AzureUp{
		azureClient:      azureClient,
		logger:           logger,
		envIDManager:     envIDManager,
		stateStore:       stateStore,
		terraformManager: terraformManager,
	}
}

func (u AzureUp) Execute(upConfig AzureUpConfig, state storage.State) error {
	u.logger.Step("verifying credentials")
	err := u.azureClient.ValidateCredentials(state.Azure.SubscriptionID, state.Azure.TenantID, state.Azure.ClientID, state.Azure.ClientSecret)

	if err != nil {
		return errors.New("Error: credentials are invalid")
	}

	if upConfig.NoDirector {
		state.NoDirector = true
	}

	state, err = u.envIDManager.Sync(state, upConfig.Name)
	if err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	state, err = u.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, u.stateStore)
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	return nil
}
