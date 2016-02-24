package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
)

type EC2SessionProvider struct {
	SessionCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Session ec2.Session
			Error   error
		}
	}
}

func (p *EC2SessionProvider) Session(config aws.Config) (ec2.Session, error) {
	p.SessionCall.Receives.Config = config

	return p.SessionCall.Returns.Session, p.SessionCall.Returns.Error
}
