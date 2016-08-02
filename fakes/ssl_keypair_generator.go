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
			CA   []byte
			Name string
		}
	}

	GenerateCACall struct {
		CallCount int

		Returns struct {
			CA    []byte
			Error error
		}

		Receives struct {
			Name string
		}
	}
}

func (c *SSLKeyPairGenerator) Generate(ca []byte, name string) (ssl.KeyPair, error) {
	c.GenerateCall.CallCount++
	c.GenerateCall.Receives.CA = ca
	c.GenerateCall.Receives.Name = name
	return c.GenerateCall.Returns.KeyPair, c.GenerateCall.Returns.Error
}

func (c *SSLKeyPairGenerator) GenerateCA(name string) ([]byte, error) {
	c.GenerateCACall.CallCount++
	c.GenerateCACall.Receives.Name = name
	return c.GenerateCACall.Returns.CA, c.GenerateCACall.Returns.Error
}
