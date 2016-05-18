package fakes

type CertificateDeleter struct {
	DeleteCall struct {
		CallCount int
		Receives  struct {
			CertificateName string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CertificateDeleter) Delete(certificateName string) error {
	c.DeleteCall.CallCount++
	c.DeleteCall.Receives.CertificateName = certificateName
	return c.DeleteCall.Returns.Error
}
