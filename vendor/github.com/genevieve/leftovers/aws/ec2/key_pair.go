package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPair struct {
	client     keyPairsClient
	name       *string
	identifier string
	rtype      string
}

func NewKeyPair(client keyPairsClient, name *string) KeyPair {
	return KeyPair{
		client:     client,
		name:       name,
		identifier: *name,
		rtype:      "EC2 Key Pair",
	}
}

func (k KeyPair) Delete() error {
	_, err := k.client.DeleteKeyPair(&awsec2.DeleteKeyPairInput{KeyName: k.name})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (k KeyPair) Name() string {
	return k.identifier
}

func (k KeyPair) Type() string {
	return k.rtype
}
