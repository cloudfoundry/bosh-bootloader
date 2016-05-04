package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type InfrastructureManager struct {
	CreateCall struct {
		CallCount int
		Stub      func(string, int, string, string, cloudformation.Client) (cloudformation.Stack, error)
		Receives  struct {
			KeyPairName               string
			StackName                 string
			LBType                    string
			LBCertificateARN          string
			CloudFormationClient      cloudformation.Client
			NumberOfAvailabilityZones int
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}

	ExistsCall struct {
		Receives struct {
			StackName string
			Client    cloudformation.Client
		}
		Returns struct {
			Exists bool
			Error  error
		}
	}

	DeleteCall struct {
		Receives struct {
			Client    cloudformation.Client
			StackName string
		}
		Returns struct {
			Error error
		}
	}

	DescribeCall struct {
		Receives struct {
			Client    cloudformation.Client
			StackName string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}
}

func (m *InfrastructureManager) Create(keyPairName string, numberOfAZs int, stackName string, lbType string, lbCertificateARN string, client cloudformation.Client) (cloudformation.Stack, error) {
	m.CreateCall.CallCount++
	m.CreateCall.Receives.StackName = stackName
	m.CreateCall.Receives.LBType = lbType
	m.CreateCall.Receives.LBCertificateARN = lbCertificateARN
	m.CreateCall.Receives.KeyPairName = keyPairName
	m.CreateCall.Receives.CloudFormationClient = client
	m.CreateCall.Receives.NumberOfAvailabilityZones = numberOfAZs

	if m.CreateCall.Stub != nil {
		return m.CreateCall.Stub(keyPairName, numberOfAZs, stackName, lbType, client)
	}

	return m.CreateCall.Returns.Stack, m.CreateCall.Returns.Error
}

func (m *InfrastructureManager) Exists(stackName string, client cloudformation.Client) (bool, error) {
	m.ExistsCall.Receives.StackName = stackName
	m.ExistsCall.Receives.Client = client

	return m.ExistsCall.Returns.Exists, m.ExistsCall.Returns.Error
}

func (m *InfrastructureManager) Delete(client cloudformation.Client, stackName string) error {
	m.DeleteCall.Receives.Client = client
	m.DeleteCall.Receives.StackName = stackName

	return m.DeleteCall.Returns.Error
}

func (m *InfrastructureManager) Describe(client cloudformation.Client, stackName string) (cloudformation.Stack, error) {
	m.DescribeCall.Receives.Client = client
	m.DescribeCall.Receives.StackName = stackName

	return m.DescribeCall.Returns.Stack, m.DescribeCall.Returns.Error
}
