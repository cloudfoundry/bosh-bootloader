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
	Name            string            `yaml:"name"`
	CloudProperties azCloudProperties `yaml:"cloud_properties"`
}

type azCloudProperties struct {
	AvailabilityZone string `yaml:"zone"`
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

func NewOpsGenerator(terraformManager terraformManager) OpsGenerator {
	return OpsGenerator{
		terraformManager: terraformManager,
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
	terraformOutputs, err := o.terraformManager.GetOutputs(state)
	if err != nil {
		return []op{}, err
	}

	var ops []op
	for i, zone := range state.GCP.Zones {
		ops = append(ops, createOp("replace", "/azs/-", az{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: azCloudProperties{
				AvailabilityZone: zone,
			},
		}))
	}

	var subnets []networkSubnet
	for i, _ := range state.GCP.Zones {
		cidr := fmt.Sprintf("10.0.%d.0/20", 16*(i+1))
		subnet, err := generateNetworkSubnet(
			fmt.Sprintf("z%d", i+1),
			cidr,
			terraformOutputs["network_name"].(string),
			terraformOutputs["subnetwork_name"].(string),
			terraformOutputs["internal_tag_name"].(string),
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
				TargetPool: terraformOutputs["concourse_target_pool"].(string),
			},
		}))
	}

	if state.LB.Type == "cf" {
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "cf-router-network-properties",
			CloudProperties: lbCloudProperties{
				BackendService: terraformOutputs["router_backend_service"].(string),
				TargetPool:     terraformOutputs["ws_target_pool"].(string),
				Tags: []string{
					terraformOutputs["router_backend_service"].(string),
					terraformOutputs["ws_target_pool"].(string),
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "diego-ssh-proxy-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: terraformOutputs["ssh_proxy_target_pool"].(string),
				Tags: []string{
					terraformOutputs["ssh_proxy_target_pool"].(string),
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "cf-tcp-router-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: terraformOutputs["tcp_router_target_pool"].(string),
				Tags: []string{
					terraformOutputs["tcp_router_target_pool"].(string),
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "credhub-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: terraformOutputs["credhub_target_pool"].(string),
				Tags: []string{
					terraformOutputs["credhub_target_pool"].(string),
				},
			},
		}))
	}

	return ops, nil
}

func generateNetworkSubnet(az, cidr, networkName, subnetworkName, internalTag string) (networkSubnet, error) {
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
			Tags:                []string{internalTag},
		},
	}, nil
}
