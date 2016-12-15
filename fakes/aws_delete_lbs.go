package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AWSDeleteLBs struct {
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}

		Returns struct {
			Error error
		}
	}
}

func (a *AWSDeleteLBs) Execute(state storage.State) error {
	a.ExecuteCall.CallCount++
	a.ExecuteCall.Receives.State = state
	return a.ExecuteCall.Returns.Error
}
