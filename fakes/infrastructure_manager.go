package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type InfrastructureManager struct {
	CreateCall struct {
		CallCount int
		Receives  struct {
			KeyPairName               string
			CloudFormationClient      cloudformation.Client
			NumberOfAvailabilityZones int
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}

	ExistsCall struct {
		Returns struct {
			Exists bool
			Error  error
		}
	}
}

func (m *InfrastructureManager) Create(keyPairName string, numberOfAZs int, client cloudformation.Client) (cloudformation.Stack, error) {
	m.CreateCall.CallCount++
	m.CreateCall.Receives.KeyPairName = keyPairName
	m.CreateCall.Receives.CloudFormationClient = client
	m.CreateCall.Receives.NumberOfAvailabilityZones = numberOfAZs
	return m.CreateCall.Returns.Stack, m.CreateCall.Returns.Error
}

func (m *InfrastructureManager) Exists(cloudformation.Client) (bool, error) {
	return m.ExistsCall.Returns.Exists, m.ExistsCall.Returns.Error
}
