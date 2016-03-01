package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type guidGenerator func() (string, error)

type KeyPairCreator struct {
	generateGUID guidGenerator
}

func NewKeyPairCreator(guidGenerator guidGenerator) KeyPairCreator {
	return KeyPairCreator{
		generateGUID: guidGenerator,
	}
}

func (c KeyPairCreator) Create(client Client) (KeyPair, error) {
	guid, err := c.generateGUID()
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
		PrivateKey: []byte(keyMaterial),
	}, nil
}
