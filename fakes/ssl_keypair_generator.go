package fakes

import "github.com/cloudfoundry/bosh-bootloader/ssl"

type SSLKeyPairGenerator struct {
	GenerateCall struct {
		CallCount int

		Returns struct {
			KeyPair ssl.KeyPair
			Error   error
		}

		Receives struct {
			CACommonName   string
			CertCommonName string
		}
	}
}

func (c *SSLKeyPairGenerator) Generate(caCommonName, certCommonName string) (ssl.KeyPair, error) {
	c.GenerateCall.CallCount++
	c.GenerateCall.Receives.CACommonName = caCommonName
	c.GenerateCall.Receives.CertCommonName = certCommonName
	return c.GenerateCall.Returns.KeyPair, c.GenerateCall.Returns.Error
}
