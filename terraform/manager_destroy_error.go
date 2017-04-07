package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerDestroyError struct {
	bblState             storage.State
	executorDestroyError executorDestroyError
}

type executorDestroyError interface {
	Error() string
	TFState() (string, error)
}

func NewManagerDestroyError(bblState storage.State, executorDestroyError executorDestroyError) ManagerDestroyError {
	return ManagerDestroyError{
		bblState:             bblState,
		executorDestroyError: executorDestroyError,
	}
}

func (m ManagerDestroyError) BBLState() (storage.State, error) {
	tfState, err := m.executorDestroyError.TFState()
	if err != nil {
		return storage.State{}, err
	}
	m.bblState.TFState = tfState
	return m.bblState, nil
}

func (m ManagerDestroyError) Error() string {
	return m.executorDestroyError.Error()
}
