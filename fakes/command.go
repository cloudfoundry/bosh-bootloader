package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type Command struct {
	CheckFastFailsCall struct {
		CallCount int
		Receives  struct {
			State           storage.State
			SubcommandFlags []string
		}
		Returns struct {
			Error error
		}
	}
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
	UsageCall struct {
		CallCount int
		Returns   struct {
			Usage string
		}
	}
}

func (c *Command) CheckFastFails(subcommandFlags []string, state storage.State) error {
	c.CheckFastFailsCall.CallCount++
	c.CheckFastFailsCall.Receives.State = state
	c.CheckFastFailsCall.Receives.SubcommandFlags = subcommandFlags

	return c.CheckFastFailsCall.Returns.Error
}

func (c *Command) Execute(subcommandFlags []string, state storage.State) error {
	c.ExecuteCall.CallCount++
	c.ExecuteCall.Receives.State = state
	c.ExecuteCall.Receives.SubcommandFlags = subcommandFlags

	return c.ExecuteCall.Returns.Error
}

func (c *Command) Usage() string {
	c.UsageCall.CallCount++
	return c.UsageCall.Returns.Usage
}
