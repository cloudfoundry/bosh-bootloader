package azure

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGenerator struct {
	terraformManager terraformManager
}

type terraformManager interface {
	GetOutputs(storage.State) (map[string]interface{}, error)
}

type op struct {
	Type  string
	Path  string
	Value interface{}
}

type network struct {
	Name    string
	Subnets []networkSubnet
	Type    string
}

type networkSubnet struct {
	AZ              string
	Gateway         string
	Range           string
	Reserved        []string
	Static          []string
	CloudProperties subnetCloudProperties `yaml:"cloud_properties"`
}

type subnetCloudProperties struct {
	VirtualNetworkName string `yaml:"virtual_network_name"`
	SubnetName         string `yaml:"subnet_name"`
	SecurityGroup      string `yaml:"security_group,omitempty"`
}

var marshal func(interface{}) ([]byte, error) = yaml.Marshal

func NewOpsGenerator(terraformManager terraformManager) OpsGenerator {
	return OpsGenerator{
		terraformManager: terraformManager,
	}
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	terraformOutputs, err := o.terraformManager.GetOutputs(state)
	if err != nil {
		return "", err
	}

	zones := []string{"z1", "z2", "z3"}
	var subnets []networkSubnet
	for i, _ := range zones {
		cidr := fmt.Sprintf("10.0.%d.0/20", 16*(i+1))
		subnet, err := generateNetworkSubnet(
			fmt.Sprintf("z%d", i+1),
			cidr,
			terraformOutputs["bosh_network_name"].(string),
			terraformOutputs["bosh_subnet_name"].(string),
			terraformOutputs["bosh_default_security_group"].(string),
		)
		if err != nil {
			panic(err)
			return "", err
		}

		subnets = append(subnets, subnet)
	}

	cloudConfigOps := []op{
		{
			Type: "replace",
			Path: "/networks/-",
			Value: network{
				Name:    "default",
				Subnets: subnets,
				Type:    "manual",
			},
		},
		{
			Type: "replace",
			Path: "/networks/-",
			Value: network{
				Name:    "private",
				Subnets: subnets,
				Type:    "manual",
			},
		},
	}

	cloudConfigOpsYAML, err := marshal(cloudConfigOps)
	if err != nil {
		return "", err
	}

	return strings.Join(
		[]string{
			BaseOps,
			string(cloudConfigOpsYAML),
		},
		"\n",
	), nil
}

func generateNetworkSubnet(az, cidr, networkName, subnetName, securityGroup string) (networkSubnet, error) {
	parsedCidr, err := bosh.ParseCIDRBlock(cidr)
	if err != nil {
		return networkSubnet{}, err
	}

	gateway := parsedCidr.GetFirstIP().Add(1).String()
	firstReserved := parsedCidr.GetFirstIP().Add(2).String()
	secondReserved := parsedCidr.GetFirstIP().Add(3).String()
	lastReserved := parsedCidr.GetLastIP().String()
	lastStatic := parsedCidr.GetLastIP().Subtract(1).String()
	firstStatic := parsedCidr.GetLastIP().Subtract(65).String()

	return networkSubnet{
		AZ:      az,
		Gateway: gateway,
		Range:   cidr,
		Reserved: []string{
			fmt.Sprintf("%s-%s", firstReserved, secondReserved),
			fmt.Sprintf("%s", lastReserved),
		},
		Static: []string{
			fmt.Sprintf("%s-%s", firstStatic, lastStatic),
		},
		CloudProperties: subnetCloudProperties{
			VirtualNetworkName: networkName,
			SubnetName:         subnetName,
			SecurityGroup:      securityGroup,
		},
	}, nil
}
