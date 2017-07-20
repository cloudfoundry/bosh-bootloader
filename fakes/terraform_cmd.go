package fakes

import (
	"errors"
	"io"
)

type TerraformCmd struct {
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

	switch args[0] {
	case "version":
		if t.RunCall.Stub != nil {
			t.RunCall.Stub(stdout)
		}
	case "init":
		t.RunCall.Initialized = true
	default:
		if !t.RunCall.Initialized {
			return errors.New("must initialize terraform v0.10.* before running any other commands")
		}
		if t.RunCall.Stub != nil {
			t.RunCall.Stub(stdout)
		}
	}

	if len(t.RunCall.Returns.Errors) >= t.RunCall.CallCount {
		return t.RunCall.Returns.Errors[t.RunCall.CallCount-1]
	}

	return nil
}
