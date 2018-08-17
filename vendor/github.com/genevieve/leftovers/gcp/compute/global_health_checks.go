package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
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

func (h GlobalHealthChecks) List(filter string) ([]common.Deletable, error) {
	checks, err := h.client.ListGlobalHealthChecks()
	if err != nil {
		return nil, fmt.Errorf("List Global Health Checks: %s", err)
	}

	var resources []common.Deletable
	for _, check := range checks.Items {
		resource := NewGlobalHealthCheck(h.client, check.Name)

		if !strings.Contains(check.Name, filter) {
			continue
		}

		proceed := h.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (h GlobalHealthChecks) Type() string {
	return "global-health-check"
}
