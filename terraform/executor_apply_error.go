package terraform

import "fmt"

type ExecutorApplyError struct {
	tfState string
	err     error
	debug   bool
}

func NewExecutorApplyError(tfState string, err error, debug bool) ExecutorApplyError {
	return ExecutorApplyError{
		tfState: tfState,
		err:     err,
		debug:   debug,
	}
}

func (t ExecutorApplyError) Error() string {
	if t.debug {
		return t.err.Error()
	} else {
		return fmt.Sprintf("%s\n%s", t.err.Error(), "use --debug for additional debug output")
	}
}

func (t ExecutorApplyError) TFState() string {
	return t.tfState
}
