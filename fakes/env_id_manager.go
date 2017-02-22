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
			EnvID string
			Error error
		}
	}
}

func (e *EnvIDManager) Sync(state storage.State, name string) (string, error) {
	e.SyncCall.CallCount++

	e.SyncCall.Receives.State = state
	e.SyncCall.Receives.Name = name
	return e.SyncCall.Returns.EnvID, e.SyncCall.Returns.Error
}
