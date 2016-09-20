package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws/iam"

type CertificateManager struct {
	CreateCall struct {
		CallCount int
		Receives  struct {
			Certificate     string
			PrivateKey      string
			Chain           string
			CertificateName string
		}
		Returns struct {
			Error error
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

func (c *CertificateManager) Create(certificate, privatekey, chain, certificateName string) error {
	c.CreateCall.CallCount++
	c.CreateCall.Receives.Certificate = certificate
	c.CreateCall.Receives.PrivateKey = privatekey
	c.CreateCall.Receives.Chain = chain
	c.CreateCall.Receives.CertificateName = certificateName

	return c.CreateCall.Returns.Error
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
