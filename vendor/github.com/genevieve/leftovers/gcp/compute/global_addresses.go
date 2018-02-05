package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type globalAddressesClient interface {
	ListGlobalAddresses() (*gcpcompute.AddressList, error)
	DeleteGlobalAddress(address string) error
}

type GlobalAddresses struct {
	client globalAddressesClient
	logger logger
}

func NewGlobalAddresses(client globalAddressesClient, logger logger) GlobalAddresses {
	return GlobalAddresses{
		client: client,
		logger: logger,
	}
}

func (a GlobalAddresses) List(filter string) ([]common.Deletable, error) {
	addresses, err := a.client.ListGlobalAddresses()
	if err != nil {
		return nil, fmt.Errorf("Listing global addresses: %s", err)
	}

	var resources []common.Deletable
	for _, address := range addresses.Items {
		resource := NewGlobalAddress(a.client, address.Name)

		if len(address.Users) > 0 {
			continue
		}

		if !strings.Contains(address.Name, filter) {
			continue
		}

		proceed := a.logger.Prompt(fmt.Sprintf("Are you sure you want to delete global address %s?", address.Name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
