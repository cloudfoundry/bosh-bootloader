package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateDeleter struct {
	client Client
}

type logger interface {
	Step(message string)
}

func NewCertificateDeleter(client Client) CertificateDeleter {
	return CertificateDeleter{
		client: client,
	}
}

func (c CertificateDeleter) Delete(certificateName string) error {
	_, err := c.client.DeleteServerCertificate(&awsiam.DeleteServerCertificateInput{
		ServerCertificateName: aws.String(certificateName),
	})
	return err
}
