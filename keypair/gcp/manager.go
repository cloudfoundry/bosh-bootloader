package gcp

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	keyPairUpdater keyPairUpdater
	keyPairDeleter keyPairDeleter
}

type keyPairUpdater interface {
	Update() (storage.KeyPair, error)
}

type keyPairDeleter interface {
	Delete(publicKey string) error
}

func NewManager(keyPairUpdater keyPairUpdater, keyPairDeleter keyPairDeleter) Manager {
	return Manager{
		keyPairUpdater: keyPairUpdater,
		keyPairDeleter: keyPairDeleter,
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

func (m Manager) Rotate(state storage.State) (storage.State, error) {
	if state.KeyPair.IsEmpty() {
		return storage.State{}, errors.New("no key found to rotate")
	}

	err := m.keyPairDeleter.Delete(state.KeyPair.PublicKey)
	if err != nil {
		return storage.State{}, err
	}

	keyPair, err := m.keyPairUpdater.Update()
	if err != nil {
		return storage.State{}, err
	}
	state.KeyPair = keyPair

	return state, nil
}
