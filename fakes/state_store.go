package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type StateStore struct {
	SetCall struct {
		CallCount int
		Receives  []SetCallReceive
		Returns   []SetCallReturn
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

type SetCallReceive struct {
	State storage.State
}

type SetCallReturn struct {
	Error error
}

func (s *StateStore) Set(state storage.State) error {
	s.SetCall.CallCount++

	s.SetCall.Receives = append(s.SetCall.Receives, SetCallReceive{State: state})

	if len(s.SetCall.Returns) < s.SetCall.CallCount {
		return nil
	}

	return s.SetCall.Returns[s.SetCall.CallCount-1].Error
}
