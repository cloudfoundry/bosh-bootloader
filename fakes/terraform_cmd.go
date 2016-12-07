package fakes

import "io"

type TerraformCmd struct {
	RunCall struct {
		Stub    func(stdout io.Writer)
		Returns struct {
			Error error
		}
		Receives struct {
			Stdout           io.Writer
			WorkingDirectory string
			Args             []string
		}
	}
}

func (t *TerraformCmd) Run(stdout io.Writer, workingDirectory string, args []string) error {
	t.RunCall.Receives.Stdout = stdout
	t.RunCall.Receives.WorkingDirectory = workingDirectory
	t.RunCall.Receives.Args = args

	if t.RunCall.Stub != nil {
		t.RunCall.Stub(stdout)
	}

	return t.RunCall.Returns.Error
}
