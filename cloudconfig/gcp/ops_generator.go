package gcp

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGenerator struct {
	terraformManager terraformManager
	zones            zones
}

type terraformManager interface {
	GetOutputs(storage.State) (map[string]interface{}, error)
}

type zones interface {
	Get(string) []string
}

type op struct {
	Type  string
	Path  string
	Value interface{}
}

type az struct {
	Name            string            `yaml:"name"`
	CloudProperties azCloudProperties `yaml:"cloud_properties"`
}

type azCloudProperties struct {
	Zone string `yaml:"zone"`
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
	EphemeralExternalIP bool   `yaml:"ephemeral_external_ip"`
	NetworkName         string `yaml:"network_name"`
	SubnetworkName      string `yaml:"subnetwork_name"`
	Tags                []string
}

type lb struct {
	Name            string
	CloudProperties lbCloudProperties `yaml:"cloud_properties"`
}

type lbCloudProperties struct {
	BackendService string   `yaml:"backend_service,omitempty"`
	TargetPool     string   `yaml:"target_pool,omitempty"`
	Tags           []string `yaml:",omitempty"`
}

var marshal func(interface{}) ([]byte, error) = yaml.Marshal

func NewOpsGenerator(terraformManager terraformManager, zones zones) OpsGenerator {
	return OpsGenerator{
		terraformManager: terraformManager,
		zones:            zones,
	}
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	ops, err := o.generateGCPOps(state)
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

func (o *OpsGenerator) generateGCPOps(state storage.State) ([]op, error) {
	var ops []op

	zones := o.zones.Get(state.GCP.Region)
	for i, zone := range zones {
		ops = append(ops, createOp("replace", "/azs/-", az{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: azCloudProperties{
				Zone: zone,
			},
		}))
	}

	outputs, err := o.terraformManager.GetOutputs(state)
	if err != nil {
		return []op{}, err
	}

	var subnets []networkSubnet
	for i, _ := range zones {
		cidr := fmt.Sprintf("10.0.%d.0/20", 16*(i+1))
		subnet, err := generateNetworkSubnet(
			fmt.Sprintf("z%d", i+1),
			cidr,
			outputs["network_name"].(string),
			outputs["subnetwork_name"].(string),
			outputs["bosh_open_tag_name"].(string),
			outputs["internal_tag_name"].(string),
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

	if state.LB.Type == "concourse" {
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "lb",
			CloudProperties: lbCloudProperties{
				TargetPool: outputs["concourse_target_pool"].(string),
			},
		}))
	}

	if state.LB.Type == "cf" {
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "cf-router-network-properties",
			CloudProperties: lbCloudProperties{
				BackendService: outputs["router_backend_service"].(string),
				TargetPool:     outputs["ws_target_pool"].(string),
				Tags: []string{
					outputs["router_backend_service"].(string),
					outputs["ws_target_pool"].(string),
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "diego-ssh-proxy-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: outputs["ssh_proxy_target_pool"].(string),
				Tags: []string{
					outputs["ssh_proxy_target_pool"].(string),
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "cf-tcp-router-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: outputs["tcp_router_target_pool"].(string),
				Tags: []string{
					outputs["tcp_router_target_pool"].(string),
				},
			},
		}))
	}

	return ops, nil
}

func generateNetworkSubnet(az, cidr, networkName, subnetworkName, boshTag, internalTag string) (networkSubnet, error) {
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
			EphemeralExternalIP: true,
			NetworkName:         networkName,
			SubnetworkName:      subnetworkName,
			Tags:                []string{boshTag, internalTag},
		},
	}, nil
}
