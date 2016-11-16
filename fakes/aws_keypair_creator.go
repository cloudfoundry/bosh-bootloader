package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws/ec2"

type KeyPairCreator struct {
	CreateCall struct {
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
		Receives struct {
			KeyPairName string
		}
	}
}

func (k *KeyPairCreator) Create(keyPairName string) (ec2.KeyPair, error) {
	k.CreateCall.Receives.KeyPairName = keyPairName
	return k.CreateCall.Returns.KeyPair, k.CreateCall.Returns.Error
}
