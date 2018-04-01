package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/common"
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

func (s ServerCertificates) ListOnly(filter string) ([]common.Deletable, error) {
	return s.getServerCertificates(filter)
}

func (s ServerCertificates) List(filter string) ([]common.Deletable, error) {
	resources, err := s.getServerCertificates(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := s.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (s ServerCertificates) getServerCertificates(filter string) ([]common.Deletable, error) {
	certificates, err := s.client.ListServerCertificates(&awsiam.ListServerCertificatesInput{})
	if err != nil {
		return nil, fmt.Errorf("List IAM Server Certificates: %s", err)
	}

	var resources []common.Deletable
	for _, c := range certificates.ServerCertificateMetadataList {
		resource := NewServerCertificate(s.client, c.ServerCertificateName)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
