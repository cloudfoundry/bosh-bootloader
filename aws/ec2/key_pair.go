package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type logger interface {
	Step(string, ...interface{})
}

type KeyPair struct {
	ec2ClientProvider ec2ClientProvider
	logger            logger
}

func NewKeyPair(ec2ClientProvider ec2ClientProvider, logger logger) KeyPair {
	return KeyPair{
		ec2ClientProvider: ec2ClientProvider,
		logger:            logger,
	}
}

func (k KeyPair) Delete(name string) error {
	k.logger.Step("deleting keypair")

	_, err := k.ec2ClientProvider.GetEC2Client().DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(name),
	})
	if err != nil {
		return err
	}

	return nil
}
