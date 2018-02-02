package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type serverCertificatesClient interface {
	ListServerCertificates(*awsiam.ListServerCertificatesInput) (*awsiam.ListServerCertificatesOutput, error)
	DeleteServerCertificate(*awsiam.DeleteServerCertificateInput) (*awsiam.DeleteServerCertificateOutput, error)
}

type ServerCertificates struct {
	client serverCertificatesClient
	logger logger
}

func NewServerCertificates(client serverCertificatesClient, logger logger) ServerCertificates {
	return ServerCertificates{
		client: client,
		logger: logger,
	}
}

func (s ServerCertificates) List(filter string) ([]common.Deletable, error) {
	certificates, err := s.client.ListServerCertificates(&awsiam.ListServerCertificatesInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing server certificates: %s", err)
	}

	var resources []common.Deletable
	for _, c := range certificates.ServerCertificateMetadataList {
		resource := NewServerCertificate(s.client, c.ServerCertificateName)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := s.logger.Prompt(fmt.Sprintf("Are you sure you want to delete server certificate %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
