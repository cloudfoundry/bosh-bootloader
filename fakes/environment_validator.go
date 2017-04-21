package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type EnvironmentValidator struct {
	ValidateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (e *EnvironmentValidator) Validate(state storage.State) error {
	e.ValidateCall.CallCount++
	e.ValidateCall.Receives.State = state
	return e.ValidateCall.Returns.Error
}
