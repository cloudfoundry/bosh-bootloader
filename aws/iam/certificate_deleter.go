package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateDeleter struct{}

func NewCertificateDeleter() CertificateDeleter {
	return CertificateDeleter{}
}

func (CertificateDeleter) Delete(certificateName string, iamClient Client) error {
	_, err := iamClient.DeleteServerCertificate(&awsiam.DeleteServerCertificateInput{
		ServerCertificateName: aws.String(certificateName),
	})
	return err
}
