package fakes

import "github.com/cloudfoundry/bosh-bootloader/certs"

type CertificateValidator struct {
	ReadAndValidateCall struct {
		CallCount int
		Returns   struct {
			CertData certs.CertData
			Error    error
		}
		Receives struct {
			Command         string
			CertificatePath string
			KeyPath         string
			ChainPath       string
		}
	}

	ReadCall struct {
		CallCount int
		Returns   struct {
			CertData certs.CertData
			Error    error
		}
		Receives struct {
			Command         string
			CertificatePath string
			KeyPath         string
			ChainPath       string
		}
	}
}

func (c *CertificateValidator) ReadAndValidate(certificatePath, keyPath, chainPath string) (certs.CertData, error) {
	c.ReadAndValidateCall.CallCount++
	c.ReadAndValidateCall.Receives.CertificatePath = certificatePath
	c.ReadAndValidateCall.Receives.KeyPath = keyPath
	c.ReadAndValidateCall.Receives.ChainPath = chainPath
	return c.ReadAndValidateCall.Returns.CertData, c.ReadAndValidateCall.Returns.Error
}

func (c *CertificateValidator) Read(certificatePath, keyPath, chainPath string) (certs.CertData, error) {
	c.ReadCall.CallCount++
	c.ReadCall.Receives.CertificatePath = certificatePath
	c.ReadCall.Receives.KeyPath = keyPath
	c.ReadCall.Receives.ChainPath = chainPath
	return c.ReadCall.Returns.CertData, c.ReadCall.Returns.Error
}
