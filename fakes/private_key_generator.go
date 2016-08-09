package fakes

import (
	"crypto/rsa"
	"io"
)

type PrivateKeyGenerator struct {
	GenerateKeyCall struct {
		CallCount int
		Receives  struct {
			Random io.Reader
			Bits   int
		}
		Returns struct {
			PrivateKey *rsa.PrivateKey
			Error      error
		}
	}
}

func (k *PrivateKeyGenerator) GenerateKey(random io.Reader, bits int) (*rsa.PrivateKey, error) {
	k.GenerateKeyCall.CallCount++
	k.GenerateKeyCall.Receives.Random = random
	k.GenerateKeyCall.Receives.Bits = bits

	return k.GenerateKeyCall.Returns.PrivateKey, k.GenerateKeyCall.Returns.Error
}
