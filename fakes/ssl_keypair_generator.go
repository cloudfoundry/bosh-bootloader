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
			Name string
		}
	}
}

func (c *SSLKeyPairGenerator) Generate(name string) (ssl.KeyPair, error) {
	c.GenerateCall.CallCount++
	c.GenerateCall.Receives.Name = name
	return c.GenerateCall.Returns.KeyPair, c.GenerateCall.Returns.Error
}
