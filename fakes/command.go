package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"
)

type Command struct {
	ExecuteCall struct {
		CallCount int
		PassState bool
		Receives  struct {
			GlobalFlags commands.GlobalFlags
			State       state.State
		}
		Returns struct {
			State state.State
			Error error
		}
	}
}

func (c *Command) Execute(globalFlags commands.GlobalFlags, s state.State) (state.State, error) {
	c.ExecuteCall.CallCount++
	c.ExecuteCall.Receives.GlobalFlags = globalFlags
	c.ExecuteCall.Receives.State = s

	if c.ExecuteCall.PassState {
		c.ExecuteCall.Returns.State = s
	}

	return c.ExecuteCall.Returns.State, c.ExecuteCall.Returns.Error
}
