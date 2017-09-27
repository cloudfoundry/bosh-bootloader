package clientmanager

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
)

type logger interface {
	Step(string, ...interface{})
}

type ClientProvider struct {
	ec2Client ec2.Client
}

func (c *ClientProvider) SetConfig(config aws.Config, logger logger) {
	c.ec2Client = ec2.NewClient(config, logger)
}

func (c *ClientProvider) Client() ec2.Client {
	return c.ec2Client
}
