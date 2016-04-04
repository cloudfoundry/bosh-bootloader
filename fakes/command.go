package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type Command struct {
	ExecuteCall struct {
		CallCount int
		PassState bool
		Receives  struct {
			GlobalFlags     commands.GlobalFlags
			State           storage.State
			SubcommandFlags []string
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
}

func (c *Command) Execute(globalFlags commands.GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	c.ExecuteCall.CallCount++
	c.ExecuteCall.Receives.GlobalFlags = globalFlags
	c.ExecuteCall.Receives.State = state
	c.ExecuteCall.Receives.SubcommandFlags = subcommandFlags

	if c.ExecuteCall.PassState {
		c.ExecuteCall.Returns.State = state
	}

	return c.ExecuteCall.Returns.State, c.ExecuteCall.Returns.Error
}
