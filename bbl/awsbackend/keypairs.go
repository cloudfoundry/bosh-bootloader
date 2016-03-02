package awsbackend

import "sync"

type KeyPair struct {
	Name string
}

type KeyPairs struct {
	mutex sync.Mutex
	store map[string]KeyPair
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

func (k *KeyPairs) All() []KeyPair {
	var keyPairs []KeyPair
	for _, keyPair := range k.store {
		keyPairs = append(keyPairs, keyPair)
	}
	return keyPairs
}
