package azure

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type OpsGenerator struct {
	terraformManager terraformManager
}

type terraformManager interface {
	GetOutputs() (terraform.Outputs, error)
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

func (o OpsGenerator) GenerateVars(state storage.State) (string, error) {
	terraformOutputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return "", err
	}

	azs := []string{"z1", "z2", "z3"}
	var varsYAML = map[string]string{
		"bosh_network_name":           terraformOutputs.GetString("bosh_network_name"),
		"bosh_subnet_name":            terraformOutputs.GetString("bosh_subnet_name"),
		"bosh_default_security_group": terraformOutputs.GetString("bosh_default_security_group"),
	}
	for i, _ := range azs {
		cidr := fmt.Sprintf("10.0.%d.0/20", 16*(i+1))
		az, err := azify(i, cidr)
		if err != nil {
			panic(err)
		}
		for name, value := range az {
			varsYAML[name] = value
		}
	}
	varsBytes, err := marshal(varsYAML)
	if err != nil {
		return "", err
	}

	return string(varsBytes), nil
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	zones := []string{"z1", "z2", "z3"}
	var subnets []networkSubnet
	for i, _ := range zones {
		subnets = append(subnets, generateNetworkSubnet(i))
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

	return strings.Join([]string{
		BaseOps,
		string(cloudConfigOpsYAML),
	}, "\n"), nil
}

func azify(az int, cidr string) (map[string]string, error) {
	parsedCidr, err := bosh.ParseCIDRBlock(cidr)
	if err != nil {
		panic(err)
	}

	gateway := parsedCidr.GetNthIP(1).String()
	firstReserved := parsedCidr.GetNthIP(2).String()
	secondReserved := parsedCidr.GetNthIP(3).String()
	lastIP := parsedCidr.GetLastIP()
	lastReserved := lastIP.String()
	lastStatic := lastIP.Subtract(1).String()
	firstStatic := lastIP.Subtract(65).String()

	azIndex := az + 1
	return map[string]string{
		fmt.Sprintf("az%d_gateway", azIndex):    gateway,
		fmt.Sprintf("az%d_range", azIndex):      cidr,
		fmt.Sprintf("az%d_reserved_1", azIndex): fmt.Sprintf("%s-%s", firstReserved, secondReserved),
		fmt.Sprintf("az%d_reserved_2", azIndex): lastReserved,
		fmt.Sprintf("az%d_static", azIndex):     fmt.Sprintf("%s-%s", firstStatic, lastStatic),
	}, nil
}

func generateNetworkSubnet(az int) networkSubnet {
	azIndex := az + 1
	return networkSubnet{
		AZ:      fmt.Sprintf("z%d", azIndex),
		Gateway: fmt.Sprintf("((az%d_gateway))", azIndex),
		Range:   fmt.Sprintf("((az%d_range))", azIndex),
		Reserved: []string{
			fmt.Sprintf("((az%d_reserved_1))", azIndex),
			fmt.Sprintf("((az%d_reserved_2))", azIndex),
		},
		Static: []string{
			fmt.Sprintf("((az%d_static))", azIndex),
		},
		CloudProperties: subnetCloudProperties{
			VirtualNetworkName: "((bosh_network_name))",
			SubnetName:         "((bosh_subnet_name))",
			SecurityGroup:      "((bosh_default_security_group))",
		},
	}
}
