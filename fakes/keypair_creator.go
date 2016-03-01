package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairCreator struct {
	CreateCall struct {
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
		Receives struct {
			Client ec2.Client
		}
	}
}

func (k *KeyPairCreator) Create(client ec2.Client) (ec2.KeyPair, error) {
	k.CreateCall.Receives.Client = client
	return k.CreateCall.Returns.KeyPair, k.CreateCall.Returns.Error
}
