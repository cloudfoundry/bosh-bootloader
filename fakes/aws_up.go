package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSUp struct {
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			AWSUpConfig commands.AWSUpConfig
			State       storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (u *AWSUp) Execute(awsUpConfig commands.AWSUpConfig, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.AWSUpConfig = awsUpConfig
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
