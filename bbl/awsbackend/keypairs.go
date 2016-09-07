package awsbackend

import (
	"sync"

	"github.com/rosenhouse/awsfaker"
)

type KeyPair struct {
	Name       string
	PrivateKey string
}

type KeyPairs struct {
	mutex         sync.Mutex
	store         map[string]KeyPair
	createKeyPair struct {
		returns struct {
			err *awsfaker.ErrorResponse
		}
	}
}

func NewKeyPairs() *KeyPairs {
	return &KeyPairs{
		store: make(map[string]KeyPair),
	}
}

func (k *KeyPairs) Set(keyPair KeyPair) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	k.store[keyPair.Name] = keyPair
}

func (k *KeyPairs) Get(name string) (KeyPair, bool) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	keyPair, ok := k.store[name]
	return keyPair, ok
}

func (k *KeyPairs) Delete(name string) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	delete(k.store, name)
}

func (k *KeyPairs) All() []KeyPair {
	var keyPairs []KeyPair
	for _, keyPair := range k.store {
		keyPairs = append(keyPairs, keyPair)
	}
	return keyPairs
}

func (k *KeyPairs) SetCreateKeyPairReturnError(err *awsfaker.ErrorResponse) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	k.createKeyPair.returns.err = err
}

func (k *KeyPairs) CreateKeyPairReturnError() *awsfaker.ErrorResponse {
	return k.createKeyPair.returns.err
}
