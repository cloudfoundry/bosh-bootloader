package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type StateStore struct {
	SetCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns []SetCallReturn
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

type SetCallReturn struct {
	Error error
}

func (s *StateStore) Set(state storage.State) error {
	s.SetCall.CallCount++
	s.SetCall.Receives.State = state

	if len(s.SetCall.Returns) < s.SetCall.CallCount {
		s.SetCall.Returns = append(s.SetCall.Returns, SetCallReturn{})
	}
	return s.SetCall.Returns[s.SetCall.CallCount-1].Error
}
