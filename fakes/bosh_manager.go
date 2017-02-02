package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type BOSHManager struct {
	CreateCall struct {
		Receives struct {
			State storage.State
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
}

func (b *BOSHManager) Create(state storage.State) (storage.State, error) {
	b.CreateCall.Receives.State = state
	state.BOSH = b.CreateCall.Returns.State.BOSH
	return state, b.CreateCall.Returns.Error
}
