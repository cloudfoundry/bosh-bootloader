package iam

import (
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

var CertificateNotFound error = errors.New("certificate not found")
var CertificateDescriptionFailure error = errors.New("failed to describe certificate")

type CertificateDescriber struct {
	client Client
}

func NewCertificateDescriber(client Client) CertificateDescriber {
	return CertificateDescriber{
		client: client,
	}
}

func (c CertificateDescriber) Describe(certificateName string) (Certificate, error) {
	output, err := c.client.GetServerCertificate(&awsiam.GetServerCertificateInput{
		ServerCertificateName: aws.String(certificateName),
	})

	if err != nil {
		if e, ok := err.(awserr.RequestFailure); ok {
			if e.StatusCode() == http.StatusNotFound && e.Code() == "NoSuchEntity" {
				return Certificate{}, CertificateNotFound
			}
		}
		return Certificate{}, err
	}

	if output.ServerCertificate == nil || output.ServerCertificate.ServerCertificateMetadata == nil {
		return Certificate{}, CertificateDescriptionFailure
	}

	return Certificate{
		Name:  aws.StringValue(output.ServerCertificate.ServerCertificateMetadata.ServerCertificateName),
		ARN:   aws.StringValue(output.ServerCertificate.ServerCertificateMetadata.Arn),
		Body:  aws.StringValue(output.ServerCertificate.CertificateBody),
		Chain: aws.StringValue(output.ServerCertificate.CertificateChain),
	}, nil
}
