package fakes

import "io"

type TerraformCmd struct {
	RunCall struct {
		CallCount int
		Stub      func(stdout io.Writer)
		Returns   struct {
			Error error
		}
		Receives struct {
			Stdout           io.Writer
			WorkingDirectory string
			Args             []string
			Debug            bool
		}
	}
}

func (t *TerraformCmd) Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error {
	t.RunCall.CallCount++
	t.RunCall.Receives.Stdout = stdout
	t.RunCall.Receives.WorkingDirectory = workingDirectory
	t.RunCall.Receives.Args = args
	t.RunCall.Receives.Debug = debug

	if t.RunCall.Stub != nil {
		t.RunCall.Stub(stdout)
	}

	return t.RunCall.Returns.Error
}
