package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type BOSHInitCommandRunner struct {
	ExecuteCall struct {
		Receives struct {
			Manifest   []byte
			PrivateKey string
			State      boshinit.State
		}
		Returns struct {
			State boshinit.State
			Error error
		}
	}
}

func (r *BOSHInitCommandRunner) Execute(manifest []byte, privateKey string, state boshinit.State) (boshinit.State, error) {
	r.ExecuteCall.Receives.Manifest = manifest
	r.ExecuteCall.Receives.PrivateKey = privateKey
	r.ExecuteCall.Receives.State = state

	return r.ExecuteCall.Returns.State, r.ExecuteCall.Returns.Error
}
