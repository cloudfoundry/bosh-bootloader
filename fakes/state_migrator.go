package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type StateMigrator struct {
	MigrateCall struct {
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

func (s *StateMigrator) Migrate(state storage.State) (storage.State, error) {
	s.MigrateCall.CallCount++
	s.MigrateCall.Receives.State = state

	return s.MigrateCall.Returns.State, s.MigrateCall.Returns.Error
}
