package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type BOSHInitRunner struct {
	DeployCall struct {
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

func (r *BOSHInitRunner) Deploy(manifest []byte, privateKey string, state boshinit.State) (boshinit.State, error) {
	r.DeployCall.Receives.Manifest = manifest
	r.DeployCall.Receives.PrivateKey = privateKey
	r.DeployCall.Receives.State = state

	return r.DeployCall.Returns.State, r.DeployCall.Returns.Error
}
