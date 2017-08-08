package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AzureUpConfig struct{}

type AzureUp struct{}

func NewAzureUp() AzureUp {
	return AzureUp{}
}

func (u AzureUp) Execute(upConfig AzureUpConfig, state storage.State) error {
	return nil
}
