package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type InfrastructureCreator struct {
	CreateCall struct {
		Receives struct {
			KeyPairName               string
			CloudFormationClient      cloudformation.Client
			NumberOfAvailabilityZones int
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}
}

func (c *InfrastructureCreator) Create(keyPairName string, numberOfAZs int, client cloudformation.Client) (cloudformation.Stack, error) {
	c.CreateCall.Receives.KeyPairName = keyPairName
	c.CreateCall.Receives.CloudFormationClient = client
	c.CreateCall.Receives.NumberOfAvailabilityZones = numberOfAZs
	return c.CreateCall.Returns.Stack, c.CreateCall.Returns.Error
}
