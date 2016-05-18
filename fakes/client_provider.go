package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/elb"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
)

type ClientProvider struct {
	CloudFormationClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client cloudformation.Client
		}
	}

	EC2ClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client ec2.Client
		}
	}

	ELBClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client elb.Client
		}
	}

	IAMClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client iam.Client
		}
	}
}

func (p *ClientProvider) CloudFormationClient(config aws.Config) cloudformation.Client {
	p.CloudFormationClientCall.Receives.Config = config

	return p.CloudFormationClientCall.Returns.Client
}

func (p *ClientProvider) EC2Client(config aws.Config) ec2.Client {
	p.EC2ClientCall.Receives.Config = config

	return p.EC2ClientCall.Returns.Client
}

func (p *ClientProvider) ELBClient(config aws.Config) elb.Client {
	p.ELBClientCall.Receives.Config = config

	return p.ELBClientCall.Returns.Client
}

func (p *ClientProvider) IAMClient(config aws.Config) iam.Client {
	p.IAMClientCall.Receives.Config = config

	return p.IAMClientCall.Returns.Client
}
