package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws/ec2"

type AWSKeyPairSynchronizer struct {
	SyncCall struct {
		CallCount int
		Receives  struct {
			KeyPair ec2.KeyPair
		}
		Returns struct {
			KeyPair ec2.KeyPair
			Error   error
		}
	}
}

func (s *AWSKeyPairSynchronizer) Sync(keyPair ec2.KeyPair) (ec2.KeyPair, error) {
	s.SyncCall.CallCount++
	s.SyncCall.Receives.KeyPair = keyPair

	return s.SyncCall.Returns.KeyPair, s.SyncCall.Returns.Error
}
