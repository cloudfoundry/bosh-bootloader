package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type azureClient interface {
	ValidateCredentials(subscriptionID, tenantID, clientID, clientSecret string) error
}

type AzureUp struct {
	azureClient azureClient
}

func NewAzureUp(azureClient azureClient) AzureUp {
	return AzureUp{
		azureClient: azureClient,
	}
}

func (u AzureUp) Execute(state storage.State) (storage.State, error) {
	err := u.azureClient.ValidateCredentials(state.Azure.SubscriptionID, state.Azure.TenantID, state.Azure.ClientID, state.Azure.ClientSecret)
	if err != nil {
		return storage.State{}, fmt.Errorf("Validate credentials: %s", err)
	}

	return state, nil
}
