package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairCreator struct {
	CreateCall struct {
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
	}
}

func (k *KeyPairCreator) Create() (ec2.KeyPair, error) {
	return k.CreateCall.Returns.KeyPair, k.CreateCall.Returns.Error
}
