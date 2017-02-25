package bosh

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerDeleteError struct {
	state storage.State
	err   error
}

func NewManagerDeleteError(state storage.State, err error) ManagerDeleteError {
	return ManagerDeleteError{
		state: state,
		err:   err,
	}
}

func (b ManagerDeleteError) Error() string {
	return b.err.Error()
}

func (b ManagerDeleteError) State() storage.State {
	return b.state
}
