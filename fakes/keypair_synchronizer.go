package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairSynchronizer struct {
	SyncCall struct {
		Receives struct {
			KeyPair ec2.KeyPair
			EnvID   string
		}
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
	}
}

func (s *KeyPairSynchronizer) Sync(keyPair ec2.KeyPair, envID string) (ec2.KeyPair, error) {
	s.SyncCall.Receives.KeyPair = keyPair
	s.SyncCall.Receives.EnvID = envID

	return s.SyncCall.Returns.KeyPair, s.SyncCall.Returns.Error
}
