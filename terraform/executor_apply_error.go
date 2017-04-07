package terraform

import (
	"fmt"
	"io/ioutil"
)

type ExecutorApplyError struct {
	tfStateFilename string
	err             error
	debug           bool
}

func NewExecutorApplyError(tfStateFilename string, err error, debug bool) ExecutorApplyError {
	return ExecutorApplyError{
		tfStateFilename: tfStateFilename,
		err:             err,
		debug:           debug,
	}
}

func (t ExecutorApplyError) Error() string {
	if t.debug {
		return t.err.Error()
	} else {
		return fmt.Sprintf("%s\n%s", t.err.Error(), "use --debug for additional debug output")
	}
}

func (t ExecutorApplyError) TFState() (string, error) {
	tfStateContents, err := ioutil.ReadFile(t.tfStateFilename)
	if err != nil {
		return "", err
	}
	return string(tfStateContents), nil
}
