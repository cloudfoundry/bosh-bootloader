package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
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
	input := &awsec2.DeleteKeyPairInput{KeyName: k.name}

	_, err := k.client.DeleteKeyPair(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "InvalidKeyPair.NotFound" {
			return nil
		}

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
