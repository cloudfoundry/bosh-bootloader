package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPLBs struct {
	Name        string
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			SubcommandFlags []string
			State           storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (u *GCPLBs) Execute(subcommandFlags []string, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.SubcommandFlags = subcommandFlags
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.Error
}
