package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type SessionProvider struct {
	SessionCall struct {
		Receives struct {
			Config ec2.Config
		}
		Returns struct {
			Session ec2.Session
			Error   error
		}
	}
}

func (p *SessionProvider) Session(config ec2.Config) (ec2.Session, error) {
	p.SessionCall.Receives.Config = config

	return p.SessionCall.Returns.Session, p.SessionCall.Returns.Error
}
