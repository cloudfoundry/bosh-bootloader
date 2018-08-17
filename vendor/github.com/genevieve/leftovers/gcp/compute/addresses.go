package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type addressesClient interface {
	ListAddresses(region string) (*gcpcompute.AddressList, error)
	DeleteAddress(region, address string) error
}

type Addresses struct {
	client  addressesClient
	logger  logger
	regions map[string]string
}

func NewAddresses(client addressesClient, logger logger, regions map[string]string) Addresses {
	return Addresses{
		client:  client,
		logger:  logger,
		regions: regions,
	}
}

func (a Addresses) List(filter string) ([]common.Deletable, error) {
	addresses := []*gcpcompute.Address{}
	for _, region := range a.regions {
		l, err := a.client.ListAddresses(region)
		if err != nil {
			return nil, fmt.Errorf("List Addresses for Region %s: %s", region, err)
		}

		addresses = append(addresses, l.Items...)
	}

	var resources []common.Deletable
	for _, address := range addresses {
		resource := NewAddress(a.client, address.Name, a.regions[address.Region], len(address.Users))

		if !strings.Contains(address.Name, filter) {
			continue
		}

		proceed := a.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (a Addresses) Type() string {
	return "address"
}
