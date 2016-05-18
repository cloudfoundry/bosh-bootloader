package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type guidGenerator interface {
	Generate() (string, error)
}

type KeyPairCreator struct {
	ec2Client     Client
	guidGenerator guidGenerator
}

func NewKeyPairCreator(ec2Client Client, guidGenerator guidGenerator) KeyPairCreator {
	return KeyPairCreator{
		ec2Client:     ec2Client,
		guidGenerator: guidGenerator,
	}
}

func (c KeyPairCreator) Create() (KeyPair, error) {
	guid, err := c.guidGenerator.Generate()
	if err != nil {
		return KeyPair{}, err
	}

	keyPairName := fmt.Sprintf("keypair-%s", guid)

	output, err := c.ec2Client.CreateKeyPair(&ec2.CreateKeyPairInput{
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
