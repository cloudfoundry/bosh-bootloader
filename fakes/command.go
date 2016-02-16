package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/commands"

type Command struct {
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			GlobalFlags commands.GlobalFlags
		}
		Returns struct {
			Error error
		}
	}
}

func (c *Command) Execute(globalFlags commands.GlobalFlags) error {
	c.ExecuteCall.CallCount++
	c.ExecuteCall.Receives.GlobalFlags = globalFlags

	return c.ExecuteCall.Returns.Error
}
