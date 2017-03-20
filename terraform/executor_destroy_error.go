package terraform

type ExecutorDestroyError struct {
	tfState       string
	internalError error
}

func NewExecutorDestroyError(tfState string, internalError error) ExecutorDestroyError {
	return ExecutorDestroyError{
		tfState:       tfState,
		internalError: internalError,
	}
}

func (e ExecutorDestroyError) Error() string {
	return e.internalError.Error()
}

func (e ExecutorDestroyError) TFState() string {
	return e.tfState
}
