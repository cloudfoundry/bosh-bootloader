package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
)

type CertificateDescriber struct {
	DescribeCall struct {
		CallCount int
		Receives  struct {
			CertificateName string
			IAMClient       iam.Client
		}
		Returns struct {
			Certificate iam.Certificate
			Error       error
		}
	}
}

func (c *CertificateDescriber) Describe(certificateName string, iamClient iam.Client) (iam.Certificate, error) {
	c.DescribeCall.CallCount++
	c.DescribeCall.Receives.CertificateName = certificateName
	c.DescribeCall.Receives.IAMClient = iamClient
	return c.DescribeCall.Returns.Certificate, c.DescribeCall.Returns.Error
}
