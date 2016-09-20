package clientmanager

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
)

type ClientProvider struct {
	ec2Client            ec2.Client
	cloudformationClient cloudformation.Client
	iamClient            iam.Client
}

func (c *ClientProvider) SetConfig(config aws.Config) {
	c.ec2Client = ec2.NewClient(config)
	c.cloudformationClient = cloudformation.NewClient(config)
	c.iamClient = iam.NewClient(config)
}

func (c *ClientProvider) GetEC2Client() ec2.Client {
	return c.ec2Client
}

func (c *ClientProvider) GetCloudFormationClient() cloudformation.Client {
	return c.cloudformationClient
}

func (c *ClientProvider) GetIAMClient() iam.Client {
	return c.iamClient
}
