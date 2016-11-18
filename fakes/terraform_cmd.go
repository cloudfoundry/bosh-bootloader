package fakes

type TerraformCmd struct {
	RunCall struct {
		Returns struct {
			Error error
		}
		Receives struct {
			Args []string
		}
	}
}

func (t *TerraformCmd) Run(args []string) error {
	t.RunCall.Receives.Args = args
	return t.RunCall.Returns.Error
}
