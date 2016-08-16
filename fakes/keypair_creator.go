package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairCreator struct {
	CreateCall struct {
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
		Receives struct {
			EnvID string
		}
	}
}

func (k *KeyPairCreator) Create(envID string) (ec2.KeyPair, error) {
	k.CreateCall.Receives.EnvID = envID
	return k.CreateCall.Returns.KeyPair, k.CreateCall.Returns.Error
}
