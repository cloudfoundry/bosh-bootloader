package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AzureUp struct {
}

func NewAzureUp() AzureUp {
	return AzureUp{}
}

func (u AzureUp) Execute(state storage.State) (storage.State, error) {
	return state, nil
}
