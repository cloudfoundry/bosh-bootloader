package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPCreateLBs struct {
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

func (u *GCPCreateLBs) Execute(args []string, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.Args = args
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
