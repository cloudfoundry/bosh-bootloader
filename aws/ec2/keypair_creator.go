package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type guidGenerator interface {
	Generate() (string, error)
}

type KeyPairCreator struct {
	guidGenerator guidGenerator
}

func NewKeyPairCreator(guidGenerator guidGenerator) KeyPairCreator {
	return KeyPairCreator{
		guidGenerator: guidGenerator,
	}
}

func (c KeyPairCreator) Create(client Client) (KeyPair, error) {
	guid, err := c.guidGenerator.Generate()
	if err != nil {
		return KeyPair{}, err
	}

	keyPairName := fmt.Sprintf("keypair-%s", guid)

	output, err := client.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: &keyPairName,
	})
	if err != nil {
		return KeyPair{}, err
	}

	var keyMaterial string
	if output.KeyMaterial != nil {
		keyMaterial = *output.KeyMaterial
	}

	return KeyPair{
		Name:       keyPairName,
		PrivateKey: keyMaterial,
	}, nil
}
