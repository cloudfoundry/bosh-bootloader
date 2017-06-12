package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type SSHKeyGetter struct {
	GetCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			PrivateKey string
			Error      error
		}
	}
}

func (s *SSHKeyGetter) Get(state storage.State) (string, error) {
	s.GetCall.CallCount++
	s.GetCall.Receives.State = state

	return s.GetCall.Returns.PrivateKey, s.GetCall.Returns.Error
}
