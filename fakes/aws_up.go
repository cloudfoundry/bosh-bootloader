package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AWSUp struct {
	Name        string
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			Args  []string
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (u *AWSUp) Execute(args []string, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.Args = args
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
