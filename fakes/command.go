package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type Command struct {
	ExecuteCall struct {
		CallCount int
		PassState bool
		Receives  struct {
			State           storage.State
			SubcommandFlags []string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *Command) Execute(subcommandFlags []string, state storage.State) error {
	c.ExecuteCall.CallCount++
	c.ExecuteCall.Receives.State = state
	c.ExecuteCall.Receives.SubcommandFlags = subcommandFlags

	return c.ExecuteCall.Returns.Error
}
