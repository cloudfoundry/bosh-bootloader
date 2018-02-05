package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type backendServicesClient interface {
	ListBackendServices() (*gcpcompute.BackendServiceList, error)
	DeleteBackendService(backendService string) error
}

type BackendServices struct {
	client backendServicesClient
	logger logger
}

func NewBackendServices(client backendServicesClient, logger logger) BackendServices {
	return BackendServices{
		client: client,
		logger: logger,
	}
}

func (b BackendServices) List(filter string) ([]common.Deletable, error) {
	backendServices, err := b.client.ListBackendServices()
	if err != nil {
		return nil, fmt.Errorf("Listing backend services: %s", err)
	}

	var resources []common.Deletable
	for _, backend := range backendServices.Items {
		resource := NewBackendService(b.client, backend.Name)

		if !strings.Contains(backend.Name, filter) {
			continue
		}

		proceed := b.logger.Prompt(fmt.Sprintf("Are you sure you want to delete backend service %s?", backend.Name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
