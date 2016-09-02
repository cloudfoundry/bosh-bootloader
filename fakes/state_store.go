package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type StateStore struct {
	SetCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}

	GetCall struct {
		CallCount int
		Receives  struct {
			Dir string
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
}

func (s *StateStore) Set(state storage.State) error {
	s.SetCall.CallCount++
	s.SetCall.Receives.State = state

	return s.SetCall.Returns.Error
}

func (s *StateStore) Get(dir string) (storage.State, error) {
	s.GetCall.Receives.Dir = dir
	s.GetCall.CallCount++

	return s.GetCall.Returns.State, s.GetCall.Returns.Error
}
