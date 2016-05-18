package fakes

type CertificateUploader struct {
	UploadCall struct {
		CallCount int
		Receives  struct {
			CertificatePath string
			PrivateKeyPath  string
		}
		Returns struct {
			CertificateName string
			Error           error
		}
	}
}

func (c *CertificateUploader) Upload(certificatePath, privateKeyPath string) (string, error) {
	c.UploadCall.CallCount++
	c.UploadCall.Receives.CertificatePath = certificatePath
	c.UploadCall.Receives.PrivateKeyPath = privateKeyPath
	return c.UploadCall.Returns.CertificateName, c.UploadCall.Returns.Error
}
