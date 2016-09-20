package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws/ec2"

type KeyPairManager struct {
	SyncCall struct {
		Receives struct {
			KeyPair ec2.KeyPair
		}
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
	}
}

func (k *KeyPairManager) Sync(keyPair ec2.KeyPair) (ec2.KeyPair, error) {
	k.SyncCall.Receives.KeyPair = keyPair

	return k.SyncCall.Returns.KeyPair, k.SyncCall.Returns.Error
}
