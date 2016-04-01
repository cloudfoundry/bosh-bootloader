package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPairDeleter struct {
	logger logger
}

func NewKeyPairDeleter(logger logger) KeyPairDeleter {
	return KeyPairDeleter{
		logger: logger,
	}
}

func (d KeyPairDeleter) Delete(client Client, name string) error {
	d.logger.Step("deleting keypair")

	_, err := client.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		return err
	}

	return nil
}
