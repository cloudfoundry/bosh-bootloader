package fakes

type CertificateValidator struct {
	ValidateCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
		Receives struct {
			Command         string
			CertificatePath string
			KeyPath         string
			ChainPath       string
		}
	}
}

func (c *CertificateValidator) Validate(command, certificatePath, keyPath, chainPath string) error {
	c.ValidateCall.CallCount++
	c.ValidateCall.Receives.Command = command
	c.ValidateCall.Receives.CertificatePath = certificatePath
	c.ValidateCall.Receives.KeyPath = keyPath
	c.ValidateCall.Receives.ChainPath = chainPath
	return c.ValidateCall.Returns.Error
}
