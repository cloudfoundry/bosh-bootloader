package iam

import (
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type Client interface {
	UploadServerCertificate(*awsiam.UploadServerCertificateInput) (*awsiam.UploadServerCertificateOutput, error)
}
