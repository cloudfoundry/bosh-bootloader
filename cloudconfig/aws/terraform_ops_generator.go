package aws

import (
	"errors"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TerraformOpsGenerator struct {
	availabilityZoneRetriever availabilityZoneRetriever
	terraformManager          terraformManager
}

type availabilityZoneRetriever interface {
	Retrieve(string) ([]string, error)
}

type terraformManager interface {
	GetOutputs(storage.State) (map[string]interface{}, error)
}

type op struct {
	Type  string
	Path  string
	Value interface{}
}

type az struct {
	Name            string
	CloudProperties azCloudProperties `yaml:"cloud_properties"`
}

type azCloudProperties struct {
	AvailabilityZone string `yaml:"availability_zone"`
}

type network struct {
	Name    string
	Type    string
	Subnets []networkSubnet
}

type networkSubnet struct {
	AZ              string
	Gateway         string
	Range           string
	Reserved        []string
	Static          []string
	CloudProperties networkSubnetCloudProperties `yaml:"cloud_properties"`
}

type networkSubnetCloudProperties struct {
	Subnet         string
	SecurityGroups []string `yaml:"security_groups"`
}

type lb struct {
	Name            string
	CloudProperties lbCloudProperties `yaml:"cloud_properties"`
}

type lbCloudProperties struct {
	ELBs           []string
	SecurityGroups []string `yaml:"security_groups"`
}

var marshal func(interface{}) ([]byte, error) = yaml.Marshal

func NewTerraformOpsGenerator(availabilityZoneRetriever availabilityZoneRetriever, terraformManager terraformManager) TerraformOpsGenerator {
	return TerraformOpsGenerator{
		availabilityZoneRetriever: availabilityZoneRetriever,
		terraformManager:          terraformManager,
	}
}

func (a TerraformOpsGenerator) Generate(state storage.State) (string, error) {
	ops, err := a.generateTerraformAWSOps(state)
	if err != nil {
		return "", err
	}

	cloudConfigOpsYAML, err := marshal(ops)
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

func createOp(opType, opPath string, value interface{}) op {
	return op{
		Type:  opType,
		Path:  opPath,
		Value: value,
	}
}

func (a TerraformOpsGenerator) generateTerraformAWSOps(state storage.State) ([]op, error) {
	azs, err := a.availabilityZoneRetriever.Retrieve(state.AWS.Region)
	if err != nil {
		return []op{}, err
	}

	ops := []op{}
	for i, awsAZ := range azs {
		op := createOp("replace", "/azs/-", az{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: azCloudProperties{
				AvailabilityZone: awsAZ,
			},
		})
		ops = append(ops, op)
	}

	terraformOutputs, err := a.terraformManager.GetOutputs(state)
	if err != nil {
		return []op{}, err
	}

	subnetCIDRs, ok := terraformOutputs["internal_subnet_cidrs"].([]interface{})
	if !ok {
		return []op{}, errors.New("missing internal subnet cidrs terraform output")
	}

	subnetNames, ok := terraformOutputs["internal_subnet_ids"].([]interface{})
	if !ok {
		return []op{}, errors.New("missing internal subnet ids terraform output")
	}

	internalSecurityGroup, ok := terraformOutputs["internal_security_group"].(string)
	if !ok {
		return []op{}, errors.New("missing internal security group terraform output")
	}

	subnets := []networkSubnet{}
	for i := range azs {
		subnet, err := generateNetworkSubnet(
			fmt.Sprintf("z%d", i+1),
			subnetCIDRs[i].(string),
			subnetNames[i].(string),
			internalSecurityGroup,
		)
		if err != nil {
			return []op{}, err
		}

		subnets = append(subnets, subnet)
	}

	ops = append(ops, createOp("replace", "/networks/-", network{
		Name:    "private",
		Subnets: subnets,
		Type:    "manual",
	}))

	ops = append(ops, createOp("replace", "/networks/-", network{
		Name:    "default",
		Subnets: subnets,
		Type:    "manual",
	}))

	return ops, nil
}

func generateNetworkSubnet(az, cidr, subnet, securityGroup string) (networkSubnet, error) {
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
		CloudProperties: networkSubnetCloudProperties{
			Subnet:         subnet,
			SecurityGroups: []string{securityGroup},
		},
	}, nil
}
