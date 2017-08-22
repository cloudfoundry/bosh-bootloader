package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type SSHKeyDeleter struct {
	DeleteCall struct {
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

func (s *SSHKeyDeleter) Delete(state storage.State) (storage.State, error) {
	s.DeleteCall.CallCount++
	s.DeleteCall.Receives.State = state

	return s.DeleteCall.Returns.State, s.DeleteCall.Returns.Error
}
