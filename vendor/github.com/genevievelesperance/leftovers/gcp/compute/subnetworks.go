package compute

import (
	"fmt"
	"strings"
	"sync"

	gcpcompute "google.golang.org/api/compute/v1"
)

type subnetworksClient interface {
	ListSubnetworks(region string) (*gcpcompute.SubnetworkList, error)
	DeleteSubnetwork(region, network string) error
}

type Subnetworks struct {
	client  subnetworksClient
	logger  logger
	regions map[string]string
}

func NewSubnetworks(client subnetworksClient, logger logger, regions map[string]string) Subnetworks {
	return Subnetworks{
		client:  client,
		logger:  logger,
		regions: regions,
	}
}

func (n Subnetworks) List(filter string) (map[string]string, error) {
	subnetworks := []*gcpcompute.Subnetwork{}
	for _, region := range n.regions {
		l, err := n.client.ListSubnetworks(region)
		if err != nil {
			return nil, fmt.Errorf("Listing subnetworks for region %s: %s", region, err)
		}

		subnetworks = append(subnetworks, l.Items...)
	}

	delete := map[string]string{}
	for _, subnetwork := range subnetworks {
		if !strings.Contains(subnetwork.Name, filter) {
			continue
		}

		proceed := n.logger.Prompt(fmt.Sprintf("Are you sure you want to delete subnetwork %s?", subnetwork.Name))
		if !proceed {
			continue
		}

		delete[subnetwork.Name] = n.regions[subnetwork.Region]
	}

	return delete, nil
}

func (n Subnetworks) Delete(subnetworks map[string]string) {
	var wg sync.WaitGroup

	for name, region := range subnetworks {
		wg.Add(1)

		go func(name string) {
			err := n.client.DeleteSubnetwork(region, name)

			if err != nil {
				n.logger.Printf("ERROR deleting subnetwork %s: %s\n", name, err)
			} else {
				n.logger.Printf("SUCCESS deleting subnetwork %s\n", name)
			}

			wg.Done()
		}(name)
	}

	wg.Wait()
}
