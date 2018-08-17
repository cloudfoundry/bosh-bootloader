package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type firewallsClient interface {
	ListFirewalls() (*gcpcompute.FirewallList, error)
	DeleteFirewall(firewall string) error
}

type Firewalls struct {
	client firewallsClient
	logger logger
}

func NewFirewalls(client firewallsClient, logger logger) Firewalls {
	return Firewalls{
		client: client,
		logger: logger,
	}
}

func (f Firewalls) List(filter string) ([]common.Deletable, error) {
	firewalls, err := f.client.ListFirewalls()
	if err != nil {
		return nil, fmt.Errorf("Listing firewalls: %s", err)
	}

	var resources []common.Deletable
	for _, firewall := range firewalls.Items {
		resource := NewFirewall(f.client, firewall.Name)

		if strings.Contains(resource.Name(), "default") {
			continue
		}

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := f.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (f Firewalls) Type() string {
	return "firewall"
}
