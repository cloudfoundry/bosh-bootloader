package fakes

import "io"

type BOSHCommand struct {
	RunCall struct {
		CallCount int
		Receives  struct {
			Stdout           io.Writer
			WorkingDirectory string
			Args             []string
		}

		Stub func(stdout io.Writer)

		Returns struct {
			Error error
		}
	}
}

func (c *BOSHCommand) Run(stdout io.Writer, workingDirectory string, args []string) error {
	c.RunCall.CallCount++
	c.RunCall.Receives.Stdout = stdout
	c.RunCall.Receives.WorkingDirectory = workingDirectory
	c.RunCall.Receives.Args = args

	if c.RunCall.Stub != nil {
		c.RunCall.Stub(stdout)
	}

	return c.RunCall.Returns.Error
}
