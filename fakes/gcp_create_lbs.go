package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPCreateLBs struct {
	Name        string
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			Config commands.GCPCreateLBsConfig
			State  storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (u *GCPCreateLBs) Execute(config commands.GCPCreateLBsConfig, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.Config = config
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
