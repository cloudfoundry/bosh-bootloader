package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairGenerator struct {
	GenerateCall struct {
		CallCount int
		Returns   struct {
			KeyPair ec2.KeyPair
			Error   error
		}
	}
}

func (g *KeyPairGenerator) Generate() (ec2.KeyPair, error) {
	g.GenerateCall.CallCount++

	return g.GenerateCall.Returns.KeyPair, g.GenerateCall.Returns.Error
}
