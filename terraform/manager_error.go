package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerError struct {
	bblState      storage.State
	executorError executorError
}

type executorError interface {
	Error() string
	TFState() (string, error)
}

func NewManagerError(bblState storage.State, executorError executorError) ManagerError {
	return ManagerError{
		bblState:      bblState,
		executorError: executorError,
	}
}

func (m ManagerError) BBLState() (storage.State, error) {
	tfState, err := m.executorError.TFState()
	if err != nil {
		return storage.State{}, err
	}
	m.bblState.TFState = tfState
	return m.bblState, nil
}

func (m ManagerError) Error() string {
	return m.executorError.Error()
}
