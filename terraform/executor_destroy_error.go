package terraform

import (
	"fmt"
	"io/ioutil"
)

type ExecutorDestroyError struct {
	tfStateFilename string
	err             error
	debug           bool
}

func NewExecutorDestroyError(tfStateFilename string, err error, debug bool) ExecutorDestroyError {
	return ExecutorDestroyError{
		tfStateFilename: tfStateFilename,
		err:             err,
		debug:           debug,
	}
}

func (t ExecutorDestroyError) Error() string {
	if t.debug {
		return t.err.Error()
	} else {
		return fmt.Sprintf("%s\n%s", t.err.Error(), "use --debug for additional debug output")
	}
}

func (t ExecutorDestroyError) TFState() (string, error) {
	tfStateContents, err := ioutil.ReadFile(t.tfStateFilename)
	if err != nil {
		return "", err
	}
	return string(tfStateContents), nil
}
