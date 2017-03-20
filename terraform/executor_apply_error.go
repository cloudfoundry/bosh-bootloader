package terraform

type ExecutorApplyError struct {
	tfState string
	err     error
}

func NewExecutorApplyError(tfState string, err error) ExecutorApplyError {
	return ExecutorApplyError{
		tfState: tfState,
		err:     err,
	}
}

func (t ExecutorApplyError) Error() string {
	return t.err.Error()
}

func (t ExecutorApplyError) TFState() string {
	return t.tfState
}
