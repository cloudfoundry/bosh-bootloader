package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerApplyError struct {
	bblState           storage.State
	executorApplyError executorApplyError
}

type executorApplyError interface {
	Error() string
}

func NewManagerApplyError(bblState storage.State, executorApplyError executorApplyError) ManagerApplyError {
	return ManagerApplyError{
		bblState:           bblState,
		executorApplyError: executorApplyError,
	}
}

func (m ManagerApplyError) BBLState() storage.State {
	return m.bblState
}

func (m ManagerApplyError) Error() string {
	return m.executorApplyError.Error()
}
