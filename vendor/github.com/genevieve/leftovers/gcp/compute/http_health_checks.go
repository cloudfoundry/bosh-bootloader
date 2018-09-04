package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type httpHealthChecksClient interface {
	ListHttpHealthChecks() ([]*gcpcompute.HttpHealthCheck, error)
	DeleteHttpHealthCheck(httpHealthCheck string) error
}

type HttpHealthChecks struct {
	client httpHealthChecksClient
	logger logger
}

func NewHttpHealthChecks(client httpHealthChecksClient, logger logger) HttpHealthChecks {
	return HttpHealthChecks{
		client: client,
		logger: logger,
	}
}

func (h HttpHealthChecks) List(filter string) ([]common.Deletable, error) {
	checks, err := h.client.ListHttpHealthChecks()
	if err != nil {
		return nil, fmt.Errorf("List Http Health Checks: %s", err)
	}

	var resources []common.Deletable
	for _, check := range checks {
		resource := NewHttpHealthCheck(h.client, check.Name)

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

func (h HttpHealthChecks) Type() string {
	return "http-health-check"
}
