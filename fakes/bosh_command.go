package fakes

import "io"

type BOSHCommand struct {
	RunCall struct {
		CallCount int
		Receives  struct {
			Stdout           io.Writer
			WorkingDirectory string
			Args             []string
			Debug            bool
		}

		Stub func()

		Returns struct {
			Error error
		}
	}
}

func (c *BOSHCommand) Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error {
	c.RunCall.CallCount++
	c.RunCall.Receives.Stdout = stdout
	c.RunCall.Receives.WorkingDirectory = workingDirectory
	c.RunCall.Receives.Args = args
	c.RunCall.Receives.Debug = debug

	if c.RunCall.Stub != nil {
		c.RunCall.Stub()
	}

	return c.RunCall.Returns.Error
}
