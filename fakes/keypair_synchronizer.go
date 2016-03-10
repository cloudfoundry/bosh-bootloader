package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
)

type KeyPairSynchronizer struct {
	SyncCall struct {
		Receives struct {
			EC2Client ec2.Client
			KeyPair   unsupported.KeyPair
		}
		Returns struct {
			KeyPair unsupported.KeyPair
			Error   error
		}
	}
}

func (s *KeyPairSynchronizer) Sync(keyPair unsupported.KeyPair, ec2Client ec2.Client) (unsupported.KeyPair, error) {
	s.SyncCall.Receives.KeyPair = keyPair
	s.SyncCall.Receives.EC2Client = ec2Client

	return s.SyncCall.Returns.KeyPair, s.SyncCall.Returns.Error
}
