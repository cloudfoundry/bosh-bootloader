package compute

import (
	"fmt"
	"strings"
	"sync"

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

func (a Addresses) List(filter string) (map[string]string, error) {
	addresses := []*gcpcompute.Address{}
	for _, region := range a.regions {
		l, err := a.client.ListAddresses(region)
		if err != nil {
			return nil, fmt.Errorf("Listing addresses for region %s: %s", region, err)
		}

		addresses = append(addresses, l.Items...)
	}

	delete := map[string]string{}
	for _, address := range addresses {
		if len(address.Users) > 0 {
			continue
		}

		if !strings.Contains(address.Name, filter) {
			continue
		}

		proceed := a.logger.Prompt(fmt.Sprintf("Are you sure you want to delete address %s?", address.Name))
		if !proceed {
			continue
		}

		delete[address.Name] = a.regions[address.Region]
	}

	return delete, nil
}

func (a Addresses) Delete(addrs map[string]string) {
	var wg sync.WaitGroup

	for name, region := range addrs {
		wg.Add(1)

		go func(name, region string) {
			err := a.client.DeleteAddress(region, name)

			if err != nil {
				a.logger.Printf("ERROR deleting address %s: %s\n", name, err)
			} else {
				a.logger.Printf("SUCCESS deleting address %s\n", name)
			}

			wg.Done()
		}(name, region)
	}

	wg.Wait()
}
