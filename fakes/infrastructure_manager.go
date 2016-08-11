package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type InfrastructureManager struct {
	CreateCall struct {
		CallCount int
		Stub      func(string, int, string, string, string) (cloudformation.Stack, error)
		Receives  struct {
			KeyPairName               string
			StackName                 string
			LBType                    string
			LBCertificateARN          string
			NumberOfAvailabilityZones int
			EnvID                     string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}

	UpdateCall struct {
		CallCount int
		Receives  struct {
			KeyPairName               string
			NumberOfAvailabilityZones int
			StackName                 string
			LBType                    string
			LBCertificateARN          string
			EnvID                     string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}

	ExistsCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Exists bool
			Error  error
		}
	}

	DeleteCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Error error
		}
	}

	DescribeCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}
}

func (m *InfrastructureManager) Create(keyPairName string, numberOfAZs int, stackName, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error) {
	m.CreateCall.CallCount++
	m.CreateCall.Receives.StackName = stackName
	m.CreateCall.Receives.LBType = lbType
	m.CreateCall.Receives.LBCertificateARN = lbCertificateARN
	m.CreateCall.Receives.KeyPairName = keyPairName
	m.CreateCall.Receives.NumberOfAvailabilityZones = numberOfAZs
	m.CreateCall.Receives.EnvID = envID

	if m.CreateCall.Stub != nil {
		return m.CreateCall.Stub(keyPairName, numberOfAZs, stackName, lbType, envID)
	}

	return m.CreateCall.Returns.Stack, m.CreateCall.Returns.Error
}

func (m *InfrastructureManager) Update(keyPairName string, numberOfAZs int, stackName, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error) {
	m.UpdateCall.CallCount++
	m.UpdateCall.Receives.KeyPairName = keyPairName
	m.UpdateCall.Receives.NumberOfAvailabilityZones = numberOfAZs
	m.UpdateCall.Receives.StackName = stackName
	m.UpdateCall.Receives.LBType = lbType
	m.UpdateCall.Receives.LBCertificateARN = lbCertificateARN
	m.UpdateCall.Receives.EnvID = envID
	return m.UpdateCall.Returns.Stack, m.UpdateCall.Returns.Error
}

func (m *InfrastructureManager) Exists(stackName string) (bool, error) {
	m.ExistsCall.Receives.StackName = stackName

	return m.ExistsCall.Returns.Exists, m.ExistsCall.Returns.Error
}

func (m *InfrastructureManager) Delete(stackName string) error {
	m.DeleteCall.Receives.StackName = stackName

	return m.DeleteCall.Returns.Error
}

func (m *InfrastructureManager) Describe(stackName string) (cloudformation.Stack, error) {
	m.DescribeCall.Receives.StackName = stackName

	return m.DescribeCall.Returns.Stack, m.DescribeCall.Returns.Error
}
