package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type JumpboxSSHKeyGetter struct {
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

func (j *JumpboxSSHKeyGetter) Get(state storage.State) (string, error) {
	j.GetCall.CallCount++
	j.GetCall.Receives.State = state

	return j.GetCall.Returns.PrivateKey, j.GetCall.Returns.Error
}
