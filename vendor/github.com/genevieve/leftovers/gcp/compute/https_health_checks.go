package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type httpsHealthChecksClient interface {
	ListHttpsHealthChecks() ([]*gcpcompute.HttpsHealthCheck, error)
	DeleteHttpsHealthCheck(httpsHealthCheck string) error
}

type HttpsHealthChecks struct {
	client httpsHealthChecksClient
	logger logger
}

func NewHttpsHealthChecks(client httpsHealthChecksClient, logger logger) HttpsHealthChecks {
	return HttpsHealthChecks{
		client: client,
		logger: logger,
	}
}

func (h HttpsHealthChecks) List(filter string) ([]common.Deletable, error) {
	checks, err := h.client.ListHttpsHealthChecks()
	if err != nil {
		return nil, fmt.Errorf("List Https Health Checks: %s", err)
	}

	var resources []common.Deletable
	for _, check := range checks {
		resource := NewHttpsHealthCheck(h.client, check.Name)

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

func (h HttpsHealthChecks) Type() string {
	return "https-health-check"
}
