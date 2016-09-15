package ec2

import "github.com/aws/aws-sdk-go/service/ec2"

type guidGenerator interface {
	Generate() (string, error)
}

type KeyPairCreator struct {
	ec2ClientProvider ec2ClientProvider
}

func NewKeyPairCreator(ec2ClientProvider ec2ClientProvider) KeyPairCreator {
	return KeyPairCreator{
		ec2ClientProvider: ec2ClientProvider,
	}
}

func (c KeyPairCreator) Create(keyPairName string) (KeyPair, error) {

	output, err := c.ec2ClientProvider.GetEC2Client().CreateKeyPair(&ec2.CreateKeyPairInput{
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
