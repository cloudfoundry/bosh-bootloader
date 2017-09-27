package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
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
}

func (c *AWSClientProvider) SetConfig(config aws.Config) {
	c.SetConfigCall.CallCount++
	c.SetConfigCall.Receives.Config = config
}

func (c *AWSClientProvider) GetEC2Client() ec2.Client {
	c.GetEC2ClientCall.CallCount++
	return c.GetEC2ClientCall.Returns.EC2Client
}
