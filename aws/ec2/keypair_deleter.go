package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPairDeleter struct {
	client Client
	logger logger
}

func NewKeyPairDeleter(client Client, logger logger) KeyPairDeleter {
	return KeyPairDeleter{
		client: client,
		logger: logger,
	}
}

func (d KeyPairDeleter) Delete(name string) error {
	d.logger.Step("deleting keypair")

	_, err := d.client.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		return err
	}

	return nil
}
