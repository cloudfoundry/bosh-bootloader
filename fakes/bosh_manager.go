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
	DeleteCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (b *BOSHManager) Create(state storage.State) (storage.State, error) {
	b.CreateCall.Receives.State = state
	state.BOSH = b.CreateCall.Returns.State.BOSH
	return state, b.CreateCall.Returns.Error
}

func (b *BOSHManager) Delete(state storage.State) error {
	b.DeleteCall.CallCount++
	b.DeleteCall.Receives.State = state
	return b.DeleteCall.Returns.Error
}
