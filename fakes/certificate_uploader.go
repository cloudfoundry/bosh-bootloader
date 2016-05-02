package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
)

type CertificateUploader struct {
	UploadCall struct {
		CallCount int
		Receives  struct {
			CertificateName string
			CertificatePath string
			PrivateKeyPath  string
			IAMClient       iam.Client
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CertificateUploader) Upload(certificateName, certificatePath, privateKeyPath string, iamClient iam.Client) error {
	c.UploadCall.CallCount++
	c.UploadCall.Receives.CertificateName = certificateName
	c.UploadCall.Receives.CertificatePath = certificatePath
	c.UploadCall.Receives.PrivateKeyPath = privateKeyPath
	c.UploadCall.Receives.IAMClient = iamClient
	return c.UploadCall.Returns.Error
}
