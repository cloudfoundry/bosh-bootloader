package fakes

import (
	"io"
)

type TerraformCLI struct {
	RunCall struct {
		CallCount int
		Stub      func(stdout io.Writer)
		Returns   struct {
			Errors []error
		}
		Initialized bool
		Receives    struct {
			Stdout           io.Writer
			WorkingDirectory string
			Args             []string
			Env              []string
		}
	}
}

func (t *TerraformCLI) RunWithEnv(stdout io.Writer, workingDirectory string, args []string, env []string) error {
	t.RunCall.CallCount++
	t.RunCall.Receives.Stdout = stdout
	t.RunCall.Receives.WorkingDirectory = workingDirectory
	t.RunCall.Receives.Args = args
	t.RunCall.Receives.Env = env

	switch args[0] {
	case "version":
		if t.RunCall.Stub != nil {
			t.RunCall.Stub(stdout)
		}
	default:
		if t.RunCall.Stub != nil {
			t.RunCall.Stub(stdout)
		}
	}

	if len(t.RunCall.Returns.Errors) >= t.RunCall.CallCount {
		return t.RunCall.Returns.Errors[t.RunCall.CallCount-1]
	}

	return nil
}

func (t *TerraformCLI) Run(stdout io.Writer, workingDirectory string, args []string) error {
	return t.RunWithEnv(stdout, workingDirectory, args, []string{})
}
