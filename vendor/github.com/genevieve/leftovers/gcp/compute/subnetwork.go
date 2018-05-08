package compute

import (
	"fmt"
	"strings"
)

type Subnetwork struct {
	client      subnetworksClient
	name        string
	clearerName string
	region      string
	kind        string
}

func NewSubnetwork(client subnetworksClient, name, region, networkUrl string) Subnetwork {
	clearerName := name
	if networkUrl != "" {
		parts := strings.Split(networkUrl, "/")
		network := parts[len(parts)-1]
		clearerName = fmt.Sprintf("%s (Network:%s)", name, network)
	}

	return Subnetwork{
		client:      client,
		name:        name,
		clearerName: clearerName,
		region:      region,
		kind:        "subnetwork",
	}
}

func (s Subnetwork) Delete() error {
	err := s.client.DeleteSubnetwork(s.region, s.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (s Subnetwork) Name() string {
	return s.clearerName
}

func (s Subnetwork) Type() string {
	return "Subnetwork"
}

func (s Subnetwork) Kind() string {
	return s.kind
}
