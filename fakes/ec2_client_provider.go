package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
)

type EC2ClientProvider struct {
	ClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client ec2.Client
			Error  error
		}
	}
}

func (p *EC2ClientProvider) Client(config aws.Config) (ec2.Client, error) {
	p.ClientCall.Receives.Config = config

	return p.ClientCall.Returns.Client, p.ClientCall.Returns.Error
}
