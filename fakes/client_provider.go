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
			Error  error
		}
	}

	EC2ClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client ec2.Client
			Error  error
		}
	}

	ELBClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client elb.Client
			Error  error
		}
	}

	IAMClientCall struct {
		Receives struct {
			Config aws.Config
		}
		Returns struct {
			Client iam.Client
			Error  error
		}
	}
}

func (p *ClientProvider) CloudFormationClient(config aws.Config) (cloudformation.Client, error) {
	p.CloudFormationClientCall.Receives.Config = config

	return p.CloudFormationClientCall.Returns.Client, p.CloudFormationClientCall.Returns.Error
}

func (p *ClientProvider) EC2Client(config aws.Config) (ec2.Client, error) {
	p.EC2ClientCall.Receives.Config = config

	return p.EC2ClientCall.Returns.Client, p.EC2ClientCall.Returns.Error
}

func (p *ClientProvider) ELBClient(config aws.Config) (elb.Client, error) {
	p.ELBClientCall.Receives.Config = config

	return p.ELBClientCall.Returns.Client, p.ELBClientCall.Returns.Error
}

func (p *ClientProvider) IAMClient(config aws.Config) (iam.Client, error) {
	p.IAMClientCall.Receives.Config = config

	return p.IAMClientCall.Returns.Client, p.IAMClientCall.Returns.Error
}
