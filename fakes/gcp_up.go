package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPUp struct {
	Name        string
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			GCPUpConfig commands.GCPUpConfig
			State       storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (u *GCPUp) Execute(gcpUpConfig commands.GCPUpConfig, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.GCPUpConfig = gcpUpConfig
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
