package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type KeyPairManager struct {
	SyncCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
	RotateCall struct {
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

func (k *KeyPairManager) Sync(state storage.State) (storage.State, error) {
	k.SyncCall.CallCount++
	k.SyncCall.Receives.State = state
	state.KeyPair = k.SyncCall.Returns.State.KeyPair
	return state, k.SyncCall.Returns.Error
}

func (k *KeyPairManager) Rotate(state storage.State) (storage.State, error) {
	k.RotateCall.CallCount++
	k.RotateCall.Receives.State = state
	state.KeyPair = k.RotateCall.Returns.State.KeyPair
	return state, k.RotateCall.Returns.Error
}
