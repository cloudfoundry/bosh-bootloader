package iam

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

//go:generate counterfeiter -o ./fakes/iam_client.go --fake-name Client . Client
type Client interface {
	GetServerCertificate(*awsiam.GetServerCertificateInput) (*awsiam.GetServerCertificateOutput, error)
	DeleteServerCertificate(*awsiam.DeleteServerCertificateInput) (*awsiam.DeleteServerCertificateOutput, error)
	DeleteUserPolicy(*awsiam.DeleteUserPolicyInput) (*awsiam.DeleteUserPolicyOutput, error)
}

func NewClient(config aws.Config) Client {
	return awsiam.New(session.New(config.ClientConfig()))
}
