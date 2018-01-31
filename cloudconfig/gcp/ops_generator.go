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

	azs, err := generateAZs(state.GCP.Zones, terraformOutputs.Map)
	if err != nil {
		return "", err
	}

	varsYAML := map[string]interface{}{}
	for k, v := range terraformOutputs.Map {
		varsYAML[k] = v
	}
	for _, az := range azs {
		for key, value := range az {
			varsYAML[key] = value
		}
	}

	varsBytes, err := marshal(varsYAML)
	if err != nil {
		return "", err
	}

	return string(varsBytes), nil
}

func generateAZs(zones []string, terraformOutputs map[string]interface{}) ([]map[string]string, error) {
	var azs []map[string]string
	for azIndex, azName := range zones {
		output := fmt.Sprintf("subnet_cidr_%d", azIndex+1)

		cidr, ok := terraformOutputs[output]
		if !ok {
			return []map[string]string{}, fmt.Errorf("Missing terraform outputs %s", output)
		}

		az, err := azify(
			azIndex+1,
			azName,
			cidr.(string),
		)

		if err != nil {
			return []map[string]string{}, err
		}

		azs = append(azs, az)
	}

	return azs, nil
}

func azify(az int, azName, cidr string) (map[string]string, error) {
	parsedCidr, err := bosh.ParseCIDRBlock(cidr)
	if err != nil {
		return map[string]string{}, err
	}

	gateway := parsedCidr.GetNthIP(1).String()
	firstReserved := parsedCidr.GetNthIP(2).String()
	secondReserved := parsedCidr.GetNthIP(3).String()
	lastReserved := parsedCidr.GetLastIP().String()
	lastStatic := parsedCidr.GetLastIP().Subtract(1).String()
	firstStatic := parsedCidr.GetLastIP().Subtract(65).String()

	return map[string]string{
		fmt.Sprintf("az%d_name", az):       azName,
		fmt.Sprintf("az%d_gateway", az):    gateway,
		fmt.Sprintf("az%d_range", az):      cidr,
		fmt.Sprintf("az%d_reserved_1", az): fmt.Sprintf("%s-%s", firstReserved, secondReserved),
		fmt.Sprintf("az%d_reserved_2", az): lastReserved,
		fmt.Sprintf("az%d_static", az):     fmt.Sprintf("%s-%s", firstStatic, lastStatic),
	}, nil
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
		subnet := generateNetworkSubnet(i)
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

func generateNetworkSubnet(az int) networkSubnet {
	az++
	return networkSubnet{
		AZ:      fmt.Sprintf("z%d", az),
		Gateway: fmt.Sprintf("((az%d_gateway))", az),
		Range:   fmt.Sprintf("((az%d_range))", az),
		Reserved: []string{
			fmt.Sprintf("((az%d_reserved_1))", az),
			fmt.Sprintf("((az%d_reserved_2))", az),
		},
		Static: []string{
			fmt.Sprintf("((az%d_static))", az),
		},
		CloudProperties: subnetCloudProperties{
			EphemeralExternalIP: true,
			NetworkName:         "((network))",
			SubnetworkName:      "((subnetwork))",
			Tags:                []string{"((internal_tag_name))"},
		},
	}
}
