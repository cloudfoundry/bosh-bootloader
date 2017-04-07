package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerApplyError struct {
	bblState           storage.State
	executorApplyError executorApplyError
}

type executorApplyError interface {
	Error() string
	TFState() (string, error)
}

func NewManagerApplyError(bblState storage.State, executorApplyError executorApplyError) ManagerApplyError {
	return ManagerApplyError{
		bblState:           bblState,
		executorApplyError: executorApplyError,
	}
}

func (m ManagerApplyError) BBLState() (storage.State, error) {
	tfState, err := m.executorApplyError.TFState()
	if err != nil {
		return storage.State{}, err
	}
	m.bblState.TFState = tfState
	return m.bblState, nil
}

func (m ManagerApplyError) Error() string {
	return m.executorApplyError.Error()
}
