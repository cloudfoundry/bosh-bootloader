package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Plan struct {
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
	ParseArgsCall struct {
		CallCount int
		Receives  struct {
			Args  []string
			State storage.State
		}
		Returns struct {
			Config commands.PlanConfig
			Error  error
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
	IsInitializedCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			IsInitialized bool
		}
	}
}

func (p *Plan) CheckFastFails(subcommandFlags []string, state storage.State) error {
	p.CheckFastFailsCall.CallCount++
	p.CheckFastFailsCall.Receives.SubcommandFlags = subcommandFlags
	p.CheckFastFailsCall.Receives.State = state

	return p.CheckFastFailsCall.Returns.Error
}

func (p *Plan) ParseArgs(args []string, state storage.State) (commands.PlanConfig, error) {
	p.ParseArgsCall.CallCount++
	p.ParseArgsCall.Receives.Args = args
	p.ParseArgsCall.Receives.State = state

	return p.ParseArgsCall.Returns.Config, p.ParseArgsCall.Returns.Error
}

func (p *Plan) Execute(args []string, state storage.State) error {
	p.ExecuteCall.CallCount++
	p.ExecuteCall.Receives.Args = args
	p.ExecuteCall.Receives.State = state

	return p.ExecuteCall.Returns.Error
}

func (p *Plan) IsInitialized(state storage.State) bool {
	p.IsInitializedCall.CallCount++
	p.IsInitializedCall.Receives.State = state

	return p.IsInitializedCall.Returns.IsInitialized
}
