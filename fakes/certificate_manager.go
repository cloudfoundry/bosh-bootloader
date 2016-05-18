package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"

type CertificateManager struct {
	CreateOrUpdateCall struct {
		CallCount int
		Receives  struct {
			Certificate string
			Name        string
			PrivateKey  string
		}
		Returns struct {
			Error           error
			CertificateName string
		}
	}

	CreateCall struct {
		CallCount int
		Receives  struct {
			Certificate string
			PrivateKey  string
		}
		Returns struct {
			CertificateName string
			Error           error
		}
	}

	DeleteCall struct {
		CallCount int
		Receives  struct {
			CertificateName string
		}
		Returns struct {
			Error error
		}
	}

	DescribeCall struct {
		CallCount int
		Stub      func(string) (iam.Certificate, error)
		Receives  struct {
			CertificateName string
		}
		Returns struct {
			Certificate iam.Certificate
			Error       error
		}
	}
}

func (c *CertificateManager) CreateOrUpdate(name, certificate, privatekey string) (string, error) {
	c.CreateOrUpdateCall.CallCount++
	c.CreateOrUpdateCall.Receives.Certificate = certificate
	c.CreateOrUpdateCall.Receives.PrivateKey = privatekey
	c.CreateOrUpdateCall.Receives.Name = name

	return c.CreateOrUpdateCall.Returns.CertificateName, c.CreateOrUpdateCall.Returns.Error
}

func (c *CertificateManager) Create(certificate, privatekey string) (string, error) {
	c.CreateCall.CallCount++
	c.CreateCall.Receives.Certificate = certificate
	c.CreateCall.Receives.PrivateKey = privatekey

	return c.CreateCall.Returns.CertificateName, c.CreateCall.Returns.Error
}

func (c *CertificateManager) Delete(certificateName string) error {
	c.DeleteCall.CallCount++
	c.DeleteCall.Receives.CertificateName = certificateName
	return c.DeleteCall.Returns.Error
}

func (c *CertificateManager) Describe(certificateName string) (iam.Certificate, error) {
	c.DescribeCall.CallCount++
	c.DescribeCall.Receives.CertificateName = certificateName

	if c.DescribeCall.Stub != nil {
		return c.DescribeCall.Stub(certificateName)
	}

	return c.DescribeCall.Returns.Certificate, c.DescribeCall.Returns.Error
}
