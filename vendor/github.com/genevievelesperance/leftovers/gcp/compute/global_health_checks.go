package compute

import (
	"fmt"
	"strings"
	"sync"

	gcpcompute "google.golang.org/api/compute/v1"
)

type globalHealthChecksClient interface {
	ListGlobalHealthChecks() (*gcpcompute.HealthCheckList, error)
	DeleteGlobalHealthCheck(globalHealthCheck string) error
}

type GlobalHealthChecks struct {
	client globalHealthChecksClient
	logger logger
}

func NewGlobalHealthChecks(client globalHealthChecksClient, logger logger) GlobalHealthChecks {
	return GlobalHealthChecks{
		client: client,
		logger: logger,
	}
}

func (h GlobalHealthChecks) List(filter string) (map[string]string, error) {
	checks, err := h.client.ListGlobalHealthChecks()
	if err != nil {
		return nil, fmt.Errorf("Listing global health checks: %s", err)
	}

	delete := map[string]string{}
	for _, check := range checks.Items {
		if !strings.Contains(check.Name, filter) {
			continue
		}

		proceed := h.logger.Prompt(fmt.Sprintf("Are you sure you want to delete global health check %s?", check.Name))
		if !proceed {
			continue
		}

		delete[check.Name] = ""
	}

	return delete, nil
}

func (h GlobalHealthChecks) Delete(checks map[string]string) {
	var wg sync.WaitGroup

	for name, _ := range checks {
		wg.Add(1)

		go func(name string) {
			err := h.client.DeleteGlobalHealthCheck(name)

			if err != nil {
				h.logger.Printf("ERROR deleting global health check %s: %s\n", name, err)
			} else {
				h.logger.Printf("SUCCESS deleting global health check %s\n", name)
			}

			wg.Done()
		}(name)
	}

	wg.Wait()
}
