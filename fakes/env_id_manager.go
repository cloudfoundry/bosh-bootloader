package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type EnvIDManager struct {
	SyncCall struct {
		CallCount int
		Receives  struct {
			State storage.State
			Name  string
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
}

func (e *EnvIDManager) Sync(state storage.State, name string) (storage.State, error) {
	e.SyncCall.CallCount++

	e.SyncCall.Receives.State = state
	e.SyncCall.Receives.Name = name
	state.EnvID = e.SyncCall.Returns.State.EnvID
	return state, e.SyncCall.Returns.Error
}
