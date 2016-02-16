package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeypairGenerator struct {
	GenerateCall struct {
		CallCount int
		Returns   struct {
			Keypair ec2.Keypair
			Error   error
		}
	}
}

func (g *KeypairGenerator) Generate() (ec2.Keypair, error) {
	g.GenerateCall.CallCount++

	return g.GenerateCall.Returns.Keypair, g.GenerateCall.Returns.Error
}
