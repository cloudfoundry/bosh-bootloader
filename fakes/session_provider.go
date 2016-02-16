package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type SessionProvider struct {
	SessionCall struct {
		Receives struct {
			Config ec2.Config
		}
		Returns struct {
			Session ec2.Session
		}
	}
}

func (p *SessionProvider) Session(config ec2.Config) ec2.Session {
	p.SessionCall.Receives.Config = config

	return p.SessionCall.Returns.Session
}
