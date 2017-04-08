package terraform

import (
	"fmt"
	"io/ioutil"
)

type ExecutorError struct {
	tfStateFilename string
	err             error
	debug           bool
}

func NewExecutorError(tfStateFilename string, err error, debug bool) ExecutorError {
	return ExecutorError{
		tfStateFilename: tfStateFilename,
		err:             err,
		debug:           debug,
	}
}

func (t ExecutorError) Error() string {
	if t.debug {
		return t.err.Error()
	} else {
		return fmt.Sprintf("%s\n%s", t.err.Error(), "use --debug for additional debug output")
	}
}

func (t ExecutorError) TFState() (string, error) {
	tfStateContents, err := ioutil.ReadFile(t.tfStateFilename)
	if err != nil {
		return "", err
	}
	return string(tfStateContents), nil
}
