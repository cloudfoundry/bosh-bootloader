package keypair

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerError struct {
	bblState      storage.State
	internalError error
}

func NewManagerError(bblState storage.State, internalError error) ManagerError {
	return ManagerError{
		bblState:      bblState,
		internalError: internalError,
	}
}

func (m ManagerError) BBLState() storage.State {
	return m.bblState
}

func (m ManagerError) Error() string {
	return m.internalError.Error()
}
