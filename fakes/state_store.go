package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/state"

type StateStore struct {
	SetCall struct {
		CallCount int
		Receives  struct {
			Dir   string
			State state.State
		}
		Returns struct {
			Error error
		}
	}

	GetCall struct {
		Receives struct {
			Dir string
		}
		Returns struct {
			State state.State
			Error error
		}
	}
}

func (s *StateStore) Set(dir string, st state.State) error {
	s.SetCall.CallCount++
	s.SetCall.Receives.Dir = dir
	s.SetCall.Receives.State = st

	return s.SetCall.Returns.Error
}

func (s *StateStore) Get(dir string) (state.State, error) {
	s.GetCall.Receives.Dir = dir

	return s.GetCall.Returns.State, s.GetCall.Returns.Error
}
