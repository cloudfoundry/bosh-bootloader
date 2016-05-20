package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateDeleter struct {
	iamClient Client
	logger    logger
}

type logger interface {
	Step(message string)
}

func NewCertificateDeleter(iamClient Client, logger logger) CertificateDeleter {
	return CertificateDeleter{
		iamClient: iamClient,
		logger:    logger,
	}
}

func (c CertificateDeleter) Delete(certificateName string) error {
	c.logger.Step("deleting certificate")

	_, err := c.iamClient.DeleteServerCertificate(&awsiam.DeleteServerCertificateInput{
		ServerCertificateName: aws.String(certificateName),
	})
	return err
}
