package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairCreator struct {
	CreateCall struct {
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
		Receives struct {
			Session ec2.Session
		}
	}
}

func (k *KeyPairCreator) Create(session ec2.Session) (ec2.KeyPair, error) {
	k.CreateCall.Receives.Session = session
	return k.CreateCall.Returns.KeyPair, k.CreateCall.Returns.Error
}
