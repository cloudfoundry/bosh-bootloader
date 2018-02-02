package compute

import (
	"fmt"
	"strings"
	"sync"

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

func (b BackendServices) List(filter string) (map[string]string, error) {
	backendServices, err := b.client.ListBackendServices()
	if err != nil {
		return nil, fmt.Errorf("Listing backend services: %s", err)
	}

	delete := map[string]string{}
	for _, backend := range backendServices.Items {
		if !strings.Contains(backend.Name, filter) {
			continue
		}

		proceed := b.logger.Prompt(fmt.Sprintf("Are you sure you want to delete backend service %s?", backend.Name))
		if !proceed {
			continue
		}

		delete[backend.Name] = ""
	}

	return delete, nil
}

func (b BackendServices) Delete(backendServices map[string]string) {
	var wg sync.WaitGroup

	for name, _ := range backendServices {
		wg.Add(1)

		go func(name string) {
			err := b.client.DeleteBackendService(name)

			if err != nil {
				b.logger.Printf("ERROR deleting backend service %s: %s\n", name, err)
			} else {
				b.logger.Printf("SUCCESS deleting backend service %s\n", name)
			}

			wg.Done()
		}(name)
	}

	wg.Wait()
}
