package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPair struct {
	client     keyPairsClient
	name       *string
	identifier string
}

func NewKeyPair(client keyPairsClient, name *string) KeyPair {
	return KeyPair{
		client:     client,
		name:       name,
		identifier: *name,
	}
}

func (k KeyPair) Delete() error {
	_, err := k.client.DeleteKeyPair(&awsec2.DeleteKeyPairInput{
		KeyName: k.name,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting key pair %s: %s", k.identifier, err)
	}

	return nil
}

func (k KeyPair) Name() string {
	return k.identifier
}

func (k KeyPair) Type() string {
	return "key pair"
}
