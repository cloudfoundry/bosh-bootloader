package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerDestroyError struct {
	bblState             storage.State
	executorDestroyError executorDestroyError
}

type executorDestroyError interface {
	Error() string
}

func NewManagerDestroyError(bblState storage.State, executorDestroyError executorDestroyError) ManagerDestroyError {
	return ManagerDestroyError{
		bblState:             bblState,
		executorDestroyError: executorDestroyError,
	}
}

func (m ManagerDestroyError) BBLState() storage.State {
	return m.bblState
}

func (m ManagerDestroyError) Error() string {
	return m.executorDestroyError.Error()
}
