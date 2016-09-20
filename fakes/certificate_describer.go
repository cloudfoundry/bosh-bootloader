package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
)

type CertificateDescriber struct {
	DescribeCall struct {
		CallCount int
		Receives  struct {
			CertificateName string
		}
		Returns struct {
			Certificate iam.Certificate
			Error       error
		}
	}
}

func (c *CertificateDescriber) Describe(certificateName string) (iam.Certificate, error) {
	c.DescribeCall.CallCount++
	c.DescribeCall.Receives.CertificateName = certificateName
	return c.DescribeCall.Returns.Certificate, c.DescribeCall.Returns.Error
}
