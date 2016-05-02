package iam

import (
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type Client interface {
	UploadServerCertificate(*awsiam.UploadServerCertificateInput) (*awsiam.UploadServerCertificateOutput, error)
	GetServerCertificate(*awsiam.GetServerCertificateInput) (*awsiam.GetServerCertificateOutput, error)
	DeleteServerCertificate(*awsiam.DeleteServerCertificateInput) (*awsiam.DeleteServerCertificateOutput, error)
}
