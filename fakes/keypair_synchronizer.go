package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairSynchronizer struct {
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

func (s *KeyPairSynchronizer) Sync(keyPair ec2.KeyPair, ec2Client ec2.Client) (ec2.KeyPair, error) {
	s.SyncCall.Receives.KeyPair = keyPair
	s.SyncCall.Receives.EC2Client = ec2Client

	return s.SyncCall.Returns.KeyPair, s.SyncCall.Returns.Error
}
