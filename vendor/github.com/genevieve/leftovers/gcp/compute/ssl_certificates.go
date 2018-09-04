package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type sslCertificatesClient interface {
	ListSslCertificates() ([]*gcpcompute.SslCertificate, error)
	DeleteSslCertificate(certificate string) error
}

type SslCertificates struct {
	client sslCertificatesClient
	logger logger
}

func NewSslCertificates(client sslCertificatesClient, logger logger) SslCertificates {
	return SslCertificates{
		client: client,
		logger: logger,
	}
}

func (s SslCertificates) List(filter string) ([]common.Deletable, error) {
	sslCertificates, err := s.client.ListSslCertificates()
	if err != nil {
		return nil, fmt.Errorf("List Ssl Certificates: %s", err)
	}

	var resources []common.Deletable
	for _, cert := range sslCertificates {
		resource := NewSslCertificate(s.client, cert.Name)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := s.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (s SslCertificates) Type() string {
	return "compute-ssl-certificate"
}
