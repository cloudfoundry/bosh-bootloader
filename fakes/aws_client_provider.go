package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSClientProvider struct {
	SetConfigCall struct {
		CallCount int
		Receives  struct {
			Config storage.AWS
		}
	}
	GetEC2ClientCall struct {
		CallCount int
		Returns   struct {
			EC2Client aws.Client
		}
	}
}

func (c *AWSClientProvider) SetConfig(config storage.AWS) {
	c.SetConfigCall.CallCount++
	c.SetConfigCall.Receives.Config = config
}

func (c *AWSClientProvider) GetEC2Client() aws.Client {
	c.GetEC2ClientCall.CallCount++
	return c.GetEC2ClientCall.Returns.EC2Client
}
