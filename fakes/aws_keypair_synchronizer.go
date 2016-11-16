package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws/ec2"

type KeyPairSynchronizer struct {
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

func (s *KeyPairSynchronizer) Sync(keyPair ec2.KeyPair) (ec2.KeyPair, error) {
	s.SyncCall.Receives.KeyPair = keyPair

	return s.SyncCall.Returns.KeyPair, s.SyncCall.Returns.Error
}
