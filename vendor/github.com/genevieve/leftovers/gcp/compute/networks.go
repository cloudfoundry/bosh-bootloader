package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
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
		return nil, fmt.Errorf("Listing networks: %s", err)
	}

	var resources []common.Deletable
	for _, network := range networks.Items {
		resource := NewNetwork(n.client, network.Name)

		if !strings.Contains(network.Name, filter) {
			continue
		}

		proceed := n.logger.Prompt(fmt.Sprintf("Are you sure you want to delete network %s?", network.Name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
