package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type Up struct {
	CheckFastFailsCall struct {
		CallCount int
		Receives  struct {
			SubcommandFlags []string
			State           storage.State
		}
		Returns struct {
			Error error
		}
	}
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

func (u *Up) CheckFastFails(subcommandFlags []string, state storage.State) error {
	u.CheckFastFailsCall.CallCount++
	u.CheckFastFailsCall.Receives.SubcommandFlags = subcommandFlags
	u.CheckFastFailsCall.Receives.State = state

	return u.CheckFastFailsCall.Returns.Error
}

func (u *Up) Execute(args []string, state storage.State) error {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.Args = args
	u.ExecuteCall.Receives.State = state

	return u.ExecuteCall.Returns.Error
}
