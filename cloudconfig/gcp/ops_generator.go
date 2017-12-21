package gcp

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

func (o OpsGenerator) GenerateVars(state storage.State) (string, error) {
	terraformOutputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return "", fmt.Errorf("Get terraform outputs: %s", err)
	}
	varsBytes, err := marshal(terraformOutputs.Map)
	if err != nil {
		panic(err) // not tested; cannot occur
	}
	return string(varsBytes), nil
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

	return strings.Join([]string{
		BaseOps,
		string(cloudConfigOpsYAML),
	}, "\n"), nil
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
			"((network_name))",
			"((subnetwork_name))",
			"((internal_tag_name))",
		)
		if err != nil {
			return []op{}, fmt.Errorf("Generating network subnet: %s", err)
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
				TargetPool: "((concourse_target_pool))",
			},
		}))
	}

	if state.LB.Type == "cf" {
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "cf-router-network-properties",
			CloudProperties: lbCloudProperties{
				BackendService: "((router_backend_service))",
				TargetPool:     "((ws_target_pool))",
				Tags: []string{
					"((router_backend_service))",
					"((ws_target_pool))",
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "diego-ssh-proxy-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: "((ssh_proxy_target_pool))",
				Tags: []string{
					"((ssh_proxy_target_pool))",
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "cf-tcp-router-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: "((tcp_router_target_pool))",
				Tags: []string{
					"((tcp_router_target_pool))",
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "credhub-network-properties",
			CloudProperties: lbCloudProperties{
				TargetPool: "((credhub_target_pool))",
				Tags: []string{
					"((credhub_target_pool))",
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

	gateway := parsedCidr.GetNthIP(1).String()
	firstReserved := parsedCidr.GetNthIP(2).String()
	secondReserved := parsedCidr.GetNthIP(3).String()
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
