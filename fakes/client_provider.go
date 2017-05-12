package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
)

type AWSClientProvider struct {
	SetConfigCall struct {
		CallCount int
		Receives  struct {
			Config aws.Config
		}
	}
	GetEC2ClientCall struct {
		CallCount int
		Returns   struct {
			EC2Client ec2.Client
		}
	}
	GetCloudFormationClientCall struct {
		CallCount int
		Returns   struct {
			CloudFormationClient cloudformation.Client
		}
	}
	GetIAMClientCall struct {
		CallCount int
		Returns   struct {
			IAMClient iam.Client
		}
	}
}

func (c *AWSClientProvider) SetConfig(config aws.Config) {
	c.SetConfigCall.CallCount++
	c.SetConfigCall.Receives.Config = config
}

func (c *AWSClientProvider) GetEC2Client() ec2.Client {
	c.GetEC2ClientCall.CallCount++
	return c.GetEC2ClientCall.Returns.EC2Client
}

func (c *AWSClientProvider) GetCloudFormationClient() cloudformation.Client {
	c.GetCloudFormationClientCall.CallCount++
	return c.GetCloudFormationClientCall.Returns.CloudFormationClient
}

func (c *AWSClientProvider) GetIAMClient() iam.Client {
	c.GetIAMClientCall.CallCount++
	return c.GetIAMClientCall.Returns.IAMClient
}
