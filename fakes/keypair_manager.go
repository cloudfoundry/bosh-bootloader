package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairManager struct {
	SyncCall struct {
		Receives struct {
			EC2Client ec2.Client
			KeyPair   ec2.KeyPair
		}
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
	}
}

func (k *KeyPairManager) Sync(ec2Client ec2.Client, keyPair ec2.KeyPair) (ec2.KeyPair, error) {
	k.SyncCall.Receives.EC2Client = ec2Client
	k.SyncCall.Receives.KeyPair = keyPair

	return k.SyncCall.Returns.KeyPair, k.SyncCall.Returns.Error
}
