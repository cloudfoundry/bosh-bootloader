package iam

import (
	"fmt"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type ServerCertificate struct {
	client     serverCertificatesClient
	name       *string
	identifier string
	rtype      string
}

func NewServerCertificate(client serverCertificatesClient, name *string) ServerCertificate {
	return ServerCertificate{
		client:     client,
		name:       name,
		identifier: *name,
		rtype:      "IAM Server Certificate",
	}
}

func (s ServerCertificate) Delete() error {
	_, err := s.client.DeleteServerCertificate(&awsiam.DeleteServerCertificateInput{
		ServerCertificateName: s.name,
	})

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (s ServerCertificate) Name() string {
	return s.identifier
}

func (s ServerCertificate) Type() string {
	return s.rtype
}
