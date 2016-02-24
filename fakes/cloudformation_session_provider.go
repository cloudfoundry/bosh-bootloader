package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
)

type CloudFormationSessionProvider struct {
	SessionCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Session cloudformation.Session
			Error   error
		}
	}
}

func (p *CloudFormationSessionProvider) Session(config aws.Config) (cloudformation.Session, error) {
	p.SessionCall.Receives.Config = config

	return p.SessionCall.Returns.Session, p.SessionCall.Returns.Error
}
