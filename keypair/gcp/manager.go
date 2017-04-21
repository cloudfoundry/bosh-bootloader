package gcp

import "github.com/cloudfoundry/bosh-bootloader/storage"

type Manager struct {
	keyPairUpdater keyPairUpdater
}

type keyPairUpdater interface {
	Update() (storage.KeyPair, error)
}

func NewManager(keyPairUpdater keyPairUpdater) Manager {
	return Manager{
		keyPairUpdater: keyPairUpdater,
	}
}

func (m Manager) Sync(state storage.State) (storage.State, error) {
	if state.KeyPair.IsEmpty() {
		keyPair, err := m.keyPairUpdater.Update()
		if err != nil {
			return storage.State{}, err
		}
		state.KeyPair = keyPair
	}

	return state, nil
}
