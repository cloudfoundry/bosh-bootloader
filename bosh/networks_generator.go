package bosh

import "fmt"

type NetworksGenerator struct {
	subnetInputs   []SubnetInput
	azAssociations map[string]string
}

type Network struct {
	Name    string          `yaml:"name"`
	Type    string          `yaml:"type"`
	Subnets []NetworkSubnet `yaml:"subnets,omitempty"`

	CloudProperties *NetworkCloudProperties `yaml:"cloud_properties,omitempty"`
}

type NetworkCloudProperties struct {
	Subnet string `yaml:"subnet"`
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
	Subnet         string   `yaml:"subnet"`
	SecurityGroups []string `yaml:"security_groups,omitempty"`
}

func NewNetworksGenerator(inputs []SubnetInput, azAssociations map[string]string) NetworksGenerator {
	return NetworksGenerator{
		subnetInputs:   inputs,
		azAssociations: azAssociations,
	}
}

func (n NetworksGenerator) Generate() ([]Network, error) {
	const MINIMUM_CIDR_SIZE = 5

	network := Network{
		Name: "private",
		Type: "manual",
	}
	for _, subnet := range n.subnetInputs {
		parsedCidr, err := ParseCIDRBlock(subnet.CIDR)
		if err != nil {
			return []Network{}, err
		}

		if parsedCidr.CIDRSize < MINIMUM_CIDR_SIZE {
			return []Network{}, fmt.Errorf(`not enough IPs allocated in CIDR block for subnet "%s"`, subnet.Subnet)
		}

		gateway := parsedCidr.GetFirstIP().Add(1).String()
		firstReserved := parsedCidr.GetFirstIP().Add(2).String()
		secondReserved := parsedCidr.GetFirstIP().Add(3).String()
		lastReserved := parsedCidr.GetLastIP().String()

		networkSubnet := NetworkSubnet{
			AZ:      n.azAssociations[subnet.AZ],
			Gateway: gateway,
			Range:   subnet.CIDR,
			Reserved: []string{
				fmt.Sprintf("%s-%s", firstReserved, secondReserved),
				fmt.Sprintf("%s", lastReserved),
			},
			CloudProperties: SubnetCloudProperties{
				Subnet:         subnet.Subnet,
				SecurityGroups: subnet.SecurityGroups,
			},
		}
		network.Subnets = append(network.Subnets, networkSubnet)
	}
	return []Network{
		network,
	}, nil
}
