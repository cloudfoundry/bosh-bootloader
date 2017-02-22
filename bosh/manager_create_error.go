package bosh

import "github.com/cloudfoundry/bosh-bootloader/storage"

type ManagerCreateError struct {
	state storage.State
	err   error
}

func NewManagerCreateError(state storage.State, err error) ManagerCreateError {
	return ManagerCreateError{
		state: state,
		err:   err,
	}
}

func (b ManagerCreateError) Error() string {
	return b.err.Error()
}

func (b ManagerCreateError) State() storage.State {
	return b.state
}
