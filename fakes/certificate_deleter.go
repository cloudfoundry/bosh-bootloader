package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
)

type CertificateDeleter struct {
	DeleteCall struct {
		CallCount int
		Receives  struct {
			CertificateName string
			IAMClient       iam.Client
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CertificateDeleter) Delete(certificateName string, iamClient iam.Client) error {
	c.DeleteCall.CallCount++
	c.DeleteCall.Receives.CertificateName = certificateName
	c.DeleteCall.Receives.IAMClient = iamClient
	return c.DeleteCall.Returns.Error
}
