package fakes

type BOSHCommand struct {
	RunCall struct {
		CallCount int
		Receives  struct {
			WorkingDirectory string
			Args             []string
			Debug            bool
		}

		Returns struct {
			Error error
		}
	}
}

func (c *BOSHCommand) Run(workingDirectory string, args []string) error {
	c.RunCall.CallCount++
	c.RunCall.Receives.WorkingDirectory = workingDirectory
	c.RunCall.Receives.Args = args

	return c.RunCall.Returns.Error
}
