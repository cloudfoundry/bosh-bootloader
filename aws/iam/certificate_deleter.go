package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateDeleter struct {
	iamClientProvider iamClientProvider
}

type logger interface {
	Step(message string)
}

func NewCertificateDeleter(iamClientProvider iamClientProvider) CertificateDeleter {
	return CertificateDeleter{
		iamClientProvider: iamClientProvider,
	}
}

func (c CertificateDeleter) Delete(certificateName string) error {
	_, err := c.iamClientProvider.GetIAMClient().DeleteServerCertificate(&awsiam.DeleteServerCertificateInput{
		ServerCertificateName: aws.String(certificateName),
	})
	return err
}
