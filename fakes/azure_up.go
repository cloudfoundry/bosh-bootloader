package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AzureUp struct {
	Name        string
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			AzureUpConfig commands.AzureUpConfig
			State         storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (u *AzureUp) Execute(azureUpConfig commands.AzureUpConfig, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.AzureUpConfig = azureUpConfig
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
