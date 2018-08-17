package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcp "google.golang.org/api/compute/v1"
)

type networksClient interface {
	ListNetworks() (*gcp.NetworkList, error)
	DeleteNetwork(network string) error
}

type Networks struct {
	client networksClient
	logger logger
}

func NewNetworks(client networksClient, logger logger) Networks {
	return Networks{
		client: client,
		logger: logger,
	}
}

func (n Networks) List(filter string) ([]common.Deletable, error) {
	networks, err := n.client.ListNetworks()
	if err != nil {
		return nil, fmt.Errorf("List Networks: %s", err)
	}

	var resources []common.Deletable
	for _, network := range networks.Items {
		resource := NewNetwork(n.client, network.Name)

		if network.Name == "default" {
			continue
		}

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := n.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (n Networks) Type() string {
	return "compute-network"
}
