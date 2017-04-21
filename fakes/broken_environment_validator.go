package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type BrokenEnvironmentValidator struct {
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

func (b *BrokenEnvironmentValidator) Validate(state storage.State) error {
	b.ValidateCall.CallCount++
	b.ValidateCall.Receives.State = state
	return b.ValidateCall.Returns.Error
}
