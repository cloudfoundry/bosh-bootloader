package fakes

type BOSHCommand struct {
	RunCall struct {
		CallCount int
		Receives  struct {
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

func (c *BOSHCommand) Run(workingDirectory string, args []string) error {
	c.RunCall.CallCount++
	c.RunCall.Receives.WorkingDirectory = workingDirectory
	c.RunCall.Receives.Args = args

	if c.RunCall.Stub != nil {
		c.RunCall.Stub()
	}

	return c.RunCall.Returns.Error
}
