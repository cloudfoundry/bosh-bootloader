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
		}
	}

	ReadAndValidatePKCS12Call struct {
		CallCount int
		Returns   struct {
			CertData certs.CertData
			Error    error
		}
		Receives struct {
			CertificatePath string
			PasswordPath    string
		}
	}

	ReadPKCS12Call struct {
		CallCount int
		Returns   struct {
			CertData certs.CertData
			Error    error
		}
		Receives struct {
			CertificatePath string
			PasswordPath    string
		}
	}
}

func (c *CertificateValidator) ReadAndValidate(certificatePath, keyPath string) (certs.CertData, error) {
	c.ReadAndValidateCall.CallCount++
	c.ReadAndValidateCall.Receives.CertificatePath = certificatePath
	c.ReadAndValidateCall.Receives.KeyPath = keyPath
	return c.ReadAndValidateCall.Returns.CertData, c.ReadAndValidateCall.Returns.Error
}

func (c *CertificateValidator) Read(certificatePath, keyPath string) (certs.CertData, error) {
	c.ReadCall.CallCount++
	c.ReadCall.Receives.CertificatePath = certificatePath
	c.ReadCall.Receives.KeyPath = keyPath
	return c.ReadCall.Returns.CertData, c.ReadCall.Returns.Error
}

func (c *CertificateValidator) ReadPKCS12(certificatePath, passwordPath string) (certs.CertData, error) {
	c.ReadPKCS12Call.CallCount++
	c.ReadPKCS12Call.Receives.CertificatePath = certificatePath
	c.ReadPKCS12Call.Receives.PasswordPath = passwordPath
	return c.ReadPKCS12Call.Returns.CertData, c.ReadPKCS12Call.Returns.Error
}

func (c *CertificateValidator) ReadAndValidatePKCS12(certificatePath, passwordPath string) (certs.CertData, error) {
	c.ReadAndValidatePKCS12Call.CallCount++
	c.ReadAndValidatePKCS12Call.Receives.CertificatePath = certificatePath
	c.ReadAndValidatePKCS12Call.Receives.PasswordPath = passwordPath
	return c.ReadAndValidatePKCS12Call.Returns.CertData, c.ReadAndValidatePKCS12Call.Returns.Error
}
