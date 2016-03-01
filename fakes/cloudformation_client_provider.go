package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
)

type CloudFormationClientProvider struct {
	ClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client cloudformation.Client
			Error  error
		}
	}
}

func (p *CloudFormationClientProvider) Client(config aws.Config) (cloudformation.Client, error) {
	p.ClientCall.Receives.Config = config

	return p.ClientCall.Returns.Client, p.ClientCall.Returns.Error
}
