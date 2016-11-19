package fakes

type TerraformCmd struct {
	RunCall struct {
		Returns struct {
			Error error
		}
		Receives struct {
			WorkingDirectory string
			Args             []string
		}
	}
}

func (t *TerraformCmd) Run(workingDirectory string, args []string) error {
	t.RunCall.Receives.WorkingDirectory = workingDirectory
	t.RunCall.Receives.Args = args
	return t.RunCall.Returns.Error
}
