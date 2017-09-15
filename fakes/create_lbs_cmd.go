package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type CreateLBsCmd struct {
	Name        string
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			Config commands.CreateLBsConfig
			State  storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (u *CreateLBsCmd) Execute(config commands.CreateLBsConfig, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.Config = config
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
