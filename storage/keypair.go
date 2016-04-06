package storage

import "reflect"

type KeyPair struct {
	Name       string `json:"name"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

func (k KeyPair) IsEmpty() bool {
	return reflect.DeepEqual(k, KeyPair{})
}
