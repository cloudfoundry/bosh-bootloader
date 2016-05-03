package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
)

type CertificateUploader struct {
	UploadCall struct {
		CallCount int
		Receives  struct {
			CertificatePath string
			PrivateKeyPath  string
			IAMClient       iam.Client
		}
		Returns struct {
			CertificateName string
			Error           error
		}
	}
}

func (c *CertificateUploader) Upload(certificatePath, privateKeyPath string, iamClient iam.Client) (string, error) {
	c.UploadCall.CallCount++
	c.UploadCall.Receives.CertificatePath = certificatePath
	c.UploadCall.Receives.PrivateKeyPath = privateKeyPath
	c.UploadCall.Receives.IAMClient = iamClient
	return c.UploadCall.Returns.CertificateName, c.UploadCall.Returns.Error
}
