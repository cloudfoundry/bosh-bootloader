package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type BOSHManager struct {
	CreateCall struct {
		CallCount int
		Receives  struct {
			State   storage.State
			OpsFile []byte
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

func (b *BOSHManager) Create(state storage.State, opsFile []byte) (storage.State, error) {
	b.CreateCall.CallCount++
	b.CreateCall.Receives.State = state
	b.CreateCall.Receives.OpsFile = opsFile
	state.BOSH = b.CreateCall.Returns.State.BOSH
	return state, b.CreateCall.Returns.Error
}

func (b *BOSHManager) Delete(state storage.State) error {
	b.DeleteCall.CallCount++
	b.DeleteCall.Receives.State = state
	return b.DeleteCall.Returns.Error
}
