package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type azureClient interface {
	ValidateCredentials(subscriptionID, tenantID, clientID, clientSecret string) error
}

type AzureUpConfig struct {
	Name string
}

type AzureUp struct {
	azureClient  azureClient
	logger       logger
	envIDManager envIDManager
	stateStore   stateStore
}

func NewAzureUp(azureClient azureClient, logger logger, envIDManager envIDManager, stateStore stateStore) AzureUp {
	return AzureUp{
		azureClient:  azureClient,
		logger:       logger,
		envIDManager: envIDManager,
		stateStore:   stateStore,
	}
}

func (u AzureUp) Execute(upConfig AzureUpConfig, state storage.State) error {
	u.logger.Step("verifying credentials")
	err := u.azureClient.ValidateCredentials(state.Azure.SubscriptionID, state.Azure.TenantID, state.Azure.ClientID, state.Azure.ClientSecret)

	if err != nil {
		return errors.New("Error: credentials are invalid")
	}

	state, err = u.envIDManager.Sync(state, upConfig.Name)
	if err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	return nil
}
