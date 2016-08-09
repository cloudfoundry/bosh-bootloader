package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

type SSLKeyPairGenerator struct {
	GenerateCall struct {
		CallCount int

		Returns struct {
			KeyPair ssl.KeyPair
			Error   error
		}

		Receives struct {
			CAData ssl.CAData
			Name   string
		}
	}

	GenerateCACall struct {
		CallCount int

		Returns struct {
			CAData ssl.CAData
			Error  error
		}

		Receives struct {
			Name string
		}
	}
}

func (c *SSLKeyPairGenerator) Generate(caData ssl.CAData, name string) (ssl.KeyPair, error) {
	c.GenerateCall.CallCount++
	c.GenerateCall.Receives.CAData = caData
	c.GenerateCall.Receives.Name = name
	return c.GenerateCall.Returns.KeyPair, c.GenerateCall.Returns.Error
}

func (c *SSLKeyPairGenerator) GenerateCA(name string) (ssl.CAData, error) {
	c.GenerateCACall.CallCount++
	c.GenerateCACall.Receives.Name = name
	return c.GenerateCACall.Returns.CAData, c.GenerateCACall.Returns.Error
}
