package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateDeleter struct {
	iamClient Client
}

func NewCertificateDeleter(iamClient Client) CertificateDeleter {
	return CertificateDeleter{
		iamClient: iamClient,
	}
}

func (c CertificateDeleter) Delete(certificateName string) error {
	_, err := c.iamClient.DeleteServerCertificate(&awsiam.DeleteServerCertificateInput{
		ServerCertificateName: aws.String(certificateName),
	})
	return err
}
