package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type UpCmd struct {
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
}

func (u *UpCmd) Execute(state storage.State) (storage.State, error) {
	u.ExecuteCall.CallCount++
	u.ExecuteCall.Receives.State = state
	return u.ExecuteCall.Returns.State, u.ExecuteCall.Returns.Error
}
