package fakes

type CertificateUploader struct {
	UploadCall struct {
		CallCount int
		Receives  struct {
			CertificatePath string
			PrivateKeyPath  string
			ChainPath       string
			CertificateName string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CertificateUploader) Upload(certificatePath, privateKeyPath, chainPath, certificateName string) error {
	c.UploadCall.CallCount++
	c.UploadCall.Receives.CertificatePath = certificatePath
	c.UploadCall.Receives.PrivateKeyPath = privateKeyPath
	c.UploadCall.Receives.ChainPath = chainPath
	c.UploadCall.Receives.CertificateName = certificateName
	return c.UploadCall.Returns.Error
}
