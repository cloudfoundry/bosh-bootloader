package azure

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

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

type lb struct {
	Name            string
	CloudProperties cloudProperties `yaml:"cloud_properties"`
}

type cloudProperties struct {
	ApplicationGateway string `yaml:"application_gateway,omitempty"`
	SecurityGroup      string `yaml:"security_group,omitempty"`
	LoadBalancer       string `yaml:"load_balancer,omitempty"`
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
	Reserved        []string
	DNS             []string
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
		return "", fmt.Errorf("Get terraform outputs: %s", err)
	}

	varsYAML := map[string]interface{}{}
	for k, v := range terraformOutputs.Map {
		varsYAML[k] = v
	}

	varsBytes, err := marshal(varsYAML)
	if err != nil {
		return "", err
	}

	return string(varsBytes), nil
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	subnet := networkSubnet{
		AZs:      []string{"z1", "z2", "z3"},
		Gateway:  "((internal_gw))",
		Range:    "((subnet_cidr))",
		Reserved: []string{"((jumpbox__internal_ip))", "((director__internal_ip))", "((internal_gw))/30"},
		DNS:      []string{"168.63.129.16"},
		CloudProperties: subnetCloudProperties{
			VirtualNetworkName: "((vnet_name))",
			SubnetName:         "((subnet_name))",
			SecurityGroup:      "((default_security_group))",
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

	switch state.LB.Type {
	case "cf":
		lbOps := []op{
			{
				Type: "replace",
				Path: "/vm_extensions/-",
				Value: lb{
					Name: "cf-router-network-properties",
					CloudProperties: cloudProperties{
						ApplicationGateway: "((cf_app_gateway_name))",
						SecurityGroup:      "((cf_security_group))",
					},
				},
			},
			{
				Type:  "replace",
				Path:  "/vm_extensions/-",
				Value: lb{Name: "cf-tcp-router-network-properties"},
			},
			{
				Type:  "replace",
				Path:  "/vm_extensions/-",
				Value: lb{Name: "diego-ssh-proxy-network-properties"},
			}}
		cloudConfigOps = append(cloudConfigOps, lbOps...)
	case "concourse":
		lbOp := op{
			Type: "replace",
			Path: "/vm_extensions/-",
			Value: lb{
				Name: "lb",
				CloudProperties: cloudProperties{
					LoadBalancer: "((concourse_lb_name))",
				},
			},
		}
		cloudConfigOps = append(cloudConfigOps, lbOp)
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
