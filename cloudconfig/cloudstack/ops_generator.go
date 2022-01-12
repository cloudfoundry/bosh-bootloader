package cloudstack

import (
	"fmt"

	"gopkg.in/yaml.v2"

	"crypto/sha1"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"strings"
)

type networkSubnet struct {
	AZS             []string                     `yaml:"azs"`
	Gateway         string                       `yaml:"gateway"`
	Range           string                       `yaml:"range"`
	Reserved        []string                     `yaml:"reserved"`
	Static          []string                     `yaml:"static"`
	DNS             string                       `yaml:"dns"`
	CloudProperties networkSubnetCloudProperties `yaml:"cloud_properties"`
}

type op struct {
	Type  string
	Path  string
	Value interface{}
}

type network struct {
	Name            string
	Type            string
	DNS             string                       `yaml:"dns,omitempty"`
	Subnets         []networkSubnet              `yaml:"subnets,omitempty"`
	CloudProperties networkSubnetCloudProperties `yaml:"cloud_properties,omitempty"`
}

func createOp(opType, opPath string, value interface{}) op {
	return op{
		Type:  opType,
		Path:  opPath,
		Value: value,
	}
}

type networkSubnetCloudProperties struct {
	Name string `yaml:"name"`
}

type OpsGenerator struct {
	terraformManager terraformManager
}

type terraformManager interface {
	GetOutputs() (terraform.Outputs, error)
}

const terraformNameCharLimit = 45

var marshal func(interface{}) ([]byte, error) = yaml.Marshal

func NewOpsGenerator(terraformManager terraformManager) OpsGenerator {
	return OpsGenerator{
		terraformManager: terraformManager,
	}
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	shortEnvID := state.EnvID
	if len(shortEnvID) > terraformNameCharLimit {
		sh1 := fmt.Sprintf("%x", sha1.Sum([]byte(state.EnvID)))
		shortEnvID = fmt.Sprintf("%s-%s", shortEnvID[:terraformNameCharLimit-8], sh1[:terraformNameCharLimit-11])
	}

	ops := []op{
		createOp("replace", "/networks/-", network{
			Name:    "default",
			Subnets: []networkSubnet{o.generateNetworkSubnet(o.generateNetworkName("control-plane", shortEnvID))},
			Type:    "manual",
		}),
		createOp("replace", "/networks/-", network{
			Name:    "control-plane",
			Subnets: []networkSubnet{o.generateNetworkSubnet(o.generateNetworkName("control-plane", shortEnvID))},
			Type:    "manual",
		}),
		createOp("replace", "/networks/-", network{
			Name:    "data-plane",
			Subnets: []networkSubnet{o.generateNetworkSubnet(o.generateNetworkName("data-plane", shortEnvID))},
			Type:    "manual",
		}),
		createOp("replace", "/networks/-", network{
			Name:    "bosh-subnet",
			Subnets: []networkSubnet{o.generateNetworkSubnet(o.generateNetworkName("bosh-subnet", shortEnvID))},
			Type:    "manual",
		}),
		createOp("replace", "/networks/-", network{
			Name:    "compilation",
			Type:    "manual",
			Subnets: []networkSubnet{o.generateNetworkSubnet(o.generateNetworkName("compilation-subnet", shortEnvID))},
		}),
	}
	if state.CloudStack.IsoSegment {
		ops = append(ops, createOp("replace", "/networks/-", network{
			Name:    "data-plane-public",
			Subnets: []networkSubnet{o.generateNetworkSubnet(o.generateNetworkName("data-plane-public", shortEnvID))},
			Type:    "manual",
		}))
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

func (o OpsGenerator) GenerateVars(state storage.State) (string, error) {

	terraformOutputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return "", fmt.Errorf("Get terraform outputs: %s", err)
	}

	requiredOutputs := []string{
		"internal_subnet_cidr_mapping",
		"internal_subnet_gw_mapping",
		"dns",
	}

	for _, output := range requiredOutputs {
		if _, ok := terraformOutputs.Map[output]; !ok {
			return "", fmt.Errorf("missing %s terraform output", output)
		}
	}
	vars := terraformOutputs.Map

	cidrMap := terraformOutputs.GetStringMap("internal_subnet_cidr_mapping")
	gwMap := terraformOutputs.GetStringMap("internal_subnet_gw_mapping")
	dns := terraformOutputs.GetStringSlice("dns")

	vars["dns"] = dns

	for networkName, cidr := range cidrMap {
		if networkName == "" {
			continue
		}
		tmpMap, err := o.generateNetworkSubnetVars(networkName, cidr, gwMap[networkName])
		if err != nil {
			return "", err
		}
		for key, val := range tmpMap {
			vars[key] = val
		}
	}

	varsBytes, err := yaml.Marshal(vars)
	if err != nil {
		panic(err) // not tested; cannot occur
	}
	return string(varsBytes), nil
}

func (o OpsGenerator) generateNetworkSubnetVars(networkName, cidr, gw string) (map[string]interface{}, error) {

	m := make(map[string]interface{})
	parsedCidr, err := bosh.ParseCIDRBlock(cidr)
	if err != nil {
		return m, err
	}

	firstReserved := parsedCidr.GetNthIP(2).String()
	secondReserved := parsedCidr.GetNthIP(6).String()
	lastStatic := parsedCidr.GetLastIP().Subtract(1).String()
	firstStatic := parsedCidr.GetLastIP().Subtract(200).String()

	m[fmt.Sprintf("gw_%s", networkName)] = gw
	m[fmt.Sprintf("cidr_%s", networkName)] = cidr
	m[fmt.Sprintf("reserved_1_%s", networkName)] = fmt.Sprintf("%s-%s", firstReserved, secondReserved)
	m[fmt.Sprintf("static_%s", networkName)] = fmt.Sprintf("%s-%s", firstStatic, lastStatic)
	return m, nil
}

func (o OpsGenerator) generateNetworkName(networkName, shortEnvID string) string {
	return fmt.Sprintf("%s-%s", shortEnvID, networkName)
}

func (o OpsGenerator) generateNetworkSubnet(networkName string) networkSubnet {
	return networkSubnet{
		AZS:     []string{"z1", "z2", "z3"},
		Gateway: fmt.Sprintf("((gw_%s))", networkName),
		Range:   fmt.Sprintf("((cidr_%s))", networkName),
		DNS:     "((dns))",
		Reserved: []string{
			fmt.Sprintf("((reserved_1_%s))", networkName),
		},
		Static: []string{
			fmt.Sprintf("((static_%s))", networkName),
		},
		CloudProperties: networkSubnetCloudProperties{
			Name: networkName,
		},
	}
}
