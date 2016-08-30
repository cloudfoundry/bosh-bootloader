package ec2

import "github.com/aws/aws-sdk-go/service/ec2"

type guidGenerator interface {
	Generate() (string, error)
}

type KeyPairCreator struct {
	ec2Client Client
}

func NewKeyPairCreator(ec2Client Client) KeyPairCreator {
	return KeyPairCreator{
		ec2Client: ec2Client,
	}
}

func (c KeyPairCreator) Create(keyPairName string) (KeyPair, error) {

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
