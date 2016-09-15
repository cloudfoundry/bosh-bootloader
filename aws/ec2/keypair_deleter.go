package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPairDeleter struct {
	ec2ClientProvider ec2ClientProvider
	logger            logger
}

func NewKeyPairDeleter(ec2ClientProvider ec2ClientProvider, logger logger) KeyPairDeleter {
	return KeyPairDeleter{
		ec2ClientProvider: ec2ClientProvider,
		logger:            logger,
	}
}

func (d KeyPairDeleter) Delete(name string) error {
	d.logger.Step("deleting keypair")

	_, err := d.ec2ClientProvider.GetEC2Client().DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		return err
	}

	return nil
}
