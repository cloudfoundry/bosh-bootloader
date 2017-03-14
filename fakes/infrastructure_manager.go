package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"

type InfrastructureManager struct {
	CreateCall struct {
		CallCount int
		Stub      func(string, []string, string, string, string) (cloudformation.Stack, error)
		Receives  struct {
			KeyPairName      string
			StackName        string
			LBType           string
			LBCertificateARN string
			AZs              []string
			BOSHAZ           string
			EnvID            string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}

	UpdateCall struct {
		CallCount int
		Receives  struct {
			KeyPairName      string
			AZs              []string
			StackName        string
			LBType           string
			LBCertificateARN string
			BOSHAZ           string
			EnvID            string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}

	ExistsCall struct {
		CallCount int
		Receives  struct {
			StackName string
		}
		Returns struct {
			Exists bool
			Error  error
		}
	}

	DeleteCall struct {
		CallCount int
		Receives  struct {
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

func (m *InfrastructureManager) Create(keyPairName string, azs []string, stackName, boshAZ, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error) {
	m.CreateCall.CallCount++
	m.CreateCall.Receives.StackName = stackName
	m.CreateCall.Receives.LBType = lbType
	m.CreateCall.Receives.LBCertificateARN = lbCertificateARN
	m.CreateCall.Receives.KeyPairName = keyPairName
	m.CreateCall.Receives.AZs = azs
	m.CreateCall.Receives.BOSHAZ = boshAZ
	m.CreateCall.Receives.EnvID = envID

	if m.CreateCall.Stub != nil {
		return m.CreateCall.Stub(keyPairName, azs, stackName, lbType, envID)
	}

	return m.CreateCall.Returns.Stack, m.CreateCall.Returns.Error
}

func (m *InfrastructureManager) Update(keyPairName string, azs []string, stackName, boshAZ, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error) {
	m.UpdateCall.CallCount++
	m.UpdateCall.Receives.KeyPairName = keyPairName
	m.UpdateCall.Receives.AZs = azs
	m.UpdateCall.Receives.StackName = stackName
	m.UpdateCall.Receives.LBType = lbType
	m.UpdateCall.Receives.LBCertificateARN = lbCertificateARN
	m.UpdateCall.Receives.BOSHAZ = boshAZ
	m.UpdateCall.Receives.EnvID = envID
	return m.UpdateCall.Returns.Stack, m.UpdateCall.Returns.Error
}

func (m *InfrastructureManager) Exists(stackName string) (bool, error) {
	m.ExistsCall.CallCount++
	m.ExistsCall.Receives.StackName = stackName

	return m.ExistsCall.Returns.Exists, m.ExistsCall.Returns.Error
}

func (m *InfrastructureManager) Delete(stackName string) error {
	m.DeleteCall.CallCount++
	m.DeleteCall.Receives.StackName = stackName

	return m.DeleteCall.Returns.Error
}

func (m *InfrastructureManager) Describe(stackName string) (cloudformation.Stack, error) {
	m.DescribeCall.Receives.StackName = stackName

	return m.DescribeCall.Returns.Stack, m.DescribeCall.Returns.Error
}
