package fakes

type CertificateUploader struct {
	UploadCall struct {
		CallCount int
		Receives  struct {
			CertificatePath string
			PrivateKeyPath  string
			ChainPath       string
		}
		Returns struct {
			CertificateName string
			Error           error
		}
	}
}

func (c *CertificateUploader) Upload(certificatePath, privateKeyPath, chainPath string) (string, error) {
	c.UploadCall.CallCount++
	c.UploadCall.Receives.CertificatePath = certificatePath
	c.UploadCall.Receives.PrivateKeyPath = privateKeyPath
	c.UploadCall.Receives.ChainPath = chainPath
	return c.UploadCall.Returns.CertificateName, c.UploadCall.Returns.Error
}
