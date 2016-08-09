package fakes

import (
	"crypto/rsa"
	"io"
)

type PrivateKeyGenerator struct {
	GenerateKeyCall struct {
		Stub      func() (*rsa.PrivateKey, error)
		CallCount int
		Receives  []GenerateKeyCallReceives
		Returns   struct {
			PrivateKey *rsa.PrivateKey
			Error      error
		}
	}
}

type GenerateKeyCallReceives struct {
	Random io.Reader
	Bits   int
}

func (k *PrivateKeyGenerator) GenerateKey(random io.Reader, bits int) (*rsa.PrivateKey, error) {
	defer func() { k.GenerateKeyCall.CallCount++ }()

	k.GenerateKeyCall.Receives = append(k.GenerateKeyCall.Receives, GenerateKeyCallReceives{
		Random: random,
		Bits:   bits,
	})

	if k.GenerateKeyCall.Stub != nil {
		return k.GenerateKeyCall.Stub()
	}

	return k.GenerateKeyCall.Returns.PrivateKey, k.GenerateKeyCall.Returns.Error
}
