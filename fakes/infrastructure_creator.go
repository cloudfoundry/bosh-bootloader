package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type InfrastructureCreator struct {
	CreateCall struct {
		Receives struct {
			KeyPairName          string
			CloudFormationClient cloudformation.Client
		}
		Returns struct {
			Error error
		}
	}
}

func (c *InfrastructureCreator) Create(keyPairName string, client cloudformation.Client) error {
	c.CreateCall.Receives.KeyPairName = keyPairName
	c.CreateCall.Receives.CloudFormationClient = client
	return c.CreateCall.Returns.Error
}
