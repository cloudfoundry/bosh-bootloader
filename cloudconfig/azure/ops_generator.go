package azure

import (
	"strings"

	yaml "gopkg.in/yaml.v2"

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
	AZs             []string
	Gateway         string
	Range           string
	CloudProperties subnetCloudProperties `yaml:"cloud_properties"`
}

type subnetCloudProperties struct {
	ResourceGroupName  string `yaml:"resource_group_name,omitempty"`
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

	subnet := networkSubnet{
		Gateway: "10.0.0.1",
		Range:   "10.0.0.0/24",
		AZs:     []string{"z1", "z2", "z3"},
		CloudProperties: subnetCloudProperties{
			ResourceGroupName:  terraformOutputs["bosh_resource_group_name"].(string),
			VirtualNetworkName: terraformOutputs["bosh_network_name"].(string),
			SubnetName:         terraformOutputs["bosh_subnet_name"].(string),
			SecurityGroup:      terraformOutputs["bosh_default_security_group"].(string),
		},
	}

	cloudConfigOps := []op{
		{
			Type: "replace",
			Path: "/networks/-",
			Value: network{
				Name:    "default",
				Subnets: []networkSubnet{subnet},
				Type:    "manual",
			},
		},
		{
			Type: "replace",
			Path: "/networks/-",
			Value: network{
				Name:    "private",
				Subnets: []networkSubnet{subnet},
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
