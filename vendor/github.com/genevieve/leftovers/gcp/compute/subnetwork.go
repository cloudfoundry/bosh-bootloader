package compute

import "fmt"

type Subnetwork struct {
	client subnetworksClient
	name   string
	region string
}

func NewSubnetwork(client subnetworksClient, name, region string) Subnetwork {
	return Subnetwork{
		client: client,
		name:   name,
		region: region,
	}
}

func (s Subnetwork) Delete() error {
	err := s.client.DeleteSubnetwork(s.region, s.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting subnetwork %s: %s", s.name, err)
	}

	return nil
}
