package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"

type CertificateUploader struct {
	UploadCall struct {
		CallCount int
		Receives  struct {
			Client      iam.Client
			Certificate string
			Name        string
			PrivateKey  string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CertificateUploader) Upload(name, certificate, privatekey string, client iam.Client) error {
	c.UploadCall.CallCount++
	c.UploadCall.Receives.Client = client
	c.UploadCall.Receives.Certificate = certificate
	c.UploadCall.Receives.PrivateKey = privatekey
	c.UploadCall.Receives.Name = name

	return c.UploadCall.Returns.Error
}
