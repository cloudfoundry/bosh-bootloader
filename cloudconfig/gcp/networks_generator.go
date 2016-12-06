package gcp

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
)

type NetworksGenerator struct {
	networkName    string
	subnetworkName string
	tags           []string
	azs            []string
}

type Network struct {
	Name    string          `yaml:"name"`
	Type    string          `yaml:"type"`
	Subnets []NetworkSubnet `yaml:"subnets"`
}

type NetworkSubnet struct {
	AZ              string                `yaml:"az"`
	Gateway         string                `yaml:"gateway"`
	Range           string                `yaml:"range"`
	Reserved        []string              `yaml:"reserved"`
	Static          []string              `yaml:"static"`
	CloudProperties SubnetCloudProperties `yaml:"cloud_properties"`
}

type SubnetCloudProperties struct {
	EphemeralExternalIP bool     `yaml:"ephemeral_external_ip"`
	NetworkName         string   `yaml:"network_name"`
	SubnetworkName      string   `yaml:"subnetwork_name"`
	Tags                []string `yaml:"tags"`
}

func NewNetworksGenerator(networkName, subnetworkName string, tags, azs []string) NetworksGenerator {
	return NetworksGenerator{
		networkName:    networkName,
		subnetworkName: subnetworkName,
		tags:           tags,
		azs:            azs,
	}
}

func (n NetworksGenerator) Generate() ([]Network, error) {
	network := Network{
		Name: "private",
		Type: "manual",
	}

	cidrBlocks := []string{}
	for i := 1; i <= len(n.azs); i++ {
		cidrBlocks = append(cidrBlocks, fmt.Sprintf("10.0.%d.0/20", 16*(i)))
	}

	for i, az := range n.azs {
		parsedCidr, err := bosh.ParseCIDRBlock(cidrBlocks[i])
		if err != nil {
			return []Network{}, err
		}

		gateway := parsedCidr.GetFirstIP().Add(1).String()
		firstReserved := parsedCidr.GetFirstIP().Add(2).String()
		secondReserved := parsedCidr.GetFirstIP().Add(3).String()
		lastReserved := parsedCidr.GetLastIP().String()
		lastStatic := parsedCidr.GetLastIP().Subtract(1).String()
		firstStatic := parsedCidr.GetLastIP().Subtract(65).String()

		networkSubnet := NetworkSubnet{
			AZ:      az,
			Gateway: gateway,
			Range:   cidrBlocks[i],
			Reserved: []string{
				fmt.Sprintf("%s-%s", firstReserved, secondReserved),
				fmt.Sprintf("%s", lastReserved),
			},
			Static: []string{
				fmt.Sprintf("%s-%s", firstStatic, lastStatic),
			},
			CloudProperties: SubnetCloudProperties{
				EphemeralExternalIP: true,
				NetworkName:         n.networkName,
				SubnetworkName:      n.subnetworkName,
				Tags:                n.tags,
			},
		}
		network.Subnets = append(network.Subnets, networkSubnet)
	}
	return []Network{
		network,
	}, nil
}
