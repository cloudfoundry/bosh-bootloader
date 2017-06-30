package iam

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type Client interface {
	GetServerCertificate(*awsiam.GetServerCertificateInput) (*awsiam.GetServerCertificateOutput, error)
	DeleteServerCertificate(*awsiam.DeleteServerCertificateInput) (*awsiam.DeleteServerCertificateOutput, error)
}

func NewClient(config aws.Config) Client {
	return awsiam.New(session.New(config.ClientConfig()))
}
