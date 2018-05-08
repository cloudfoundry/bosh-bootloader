package aws

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type OpsGenerator struct {
	terraformManager  terraformManager
	availabilityZones availabilityZones
}

type availabilityZones interface {
	RetrieveAZs(region string) ([]string, error)
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
	ELBs           []string `yaml:"elbs,omitempty"`
	LBTargetGroups string   `yaml:"lb_target_groups,omitempty"`
	SecurityGroups []string `yaml:"security_groups"`
}

var marshal func(interface{}) ([]byte, error) = yaml.Marshal

func NewOpsGenerator(terraformManager terraformManager, availabilityZones availabilityZones) OpsGenerator {
	return OpsGenerator{
		terraformManager:  terraformManager,
		availabilityZones: availabilityZones,
	}
}

func (o OpsGenerator) GenerateVars(state storage.State) (string, error) {
	terraformOutputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return "", fmt.Errorf("Get terraform outputs: %s", err)
	}

	requiredOutputs := []string{
		"internal_security_group",
		"internal_az_subnet_id_mapping",
		"internal_az_subnet_cidr_mapping",
	}
	switch state.LB.Type {
	case "concourse":
		requiredOutputs = append(
			requiredOutputs,
			"concourse_lb_target_groups",
			"concourse_lb_internal_security_group",
		)
	case "cf":
		requiredOutputs = append(
			requiredOutputs,
			"cf_router_lb_name",
			"cf_router_lb_internal_security_group",
			"cf_ssh_lb_name",
			"cf_ssh_lb_internal_security_group",
			"cf_tcp_lb_name",
			"cf_tcp_lb_internal_security_group",
		)
	}

	for _, output := range requiredOutputs {
		if _, ok := terraformOutputs.Map[output]; !ok {
			return "", fmt.Errorf("missing %s terraform output", output)
		}
	}

	internalAZSubnetIDMap := terraformOutputs.GetStringMap("internal_az_subnet_id_mapping")
	internalAZSubnetCIDRMap := terraformOutputs.GetStringMap("internal_az_subnet_cidr_mapping")

	azs, err := generateAZs(0, internalAZSubnetIDMap, internalAZSubnetCIDRMap)
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

	isoSegAZSubnetIDMap := terraformOutputs.GetStringMap("iso_az_subnet_id_mapping")
	isoSegAZSubnetCIDRMap := terraformOutputs.GetStringMap("iso_az_subnet_cidr_mapping")
	if len(isoSegAZSubnetIDMap) > 0 && len(isoSegAZSubnetCIDRMap) > 0 {
		isoSegAzs, err := generateAZs(len(azs), isoSegAZSubnetIDMap, isoSegAZSubnetCIDRMap)
		if err == nil {
			for _, az := range isoSegAzs {
				for key, value := range az {
					varsYAML[key] = value
				}
			}
		}
	}

	varsBytes, err := marshal(varsYAML)
	if err != nil {
		panic(err) // not tested; cannot occur
	}
	return string(varsBytes), nil
}

func generateAZs(startingIndex int, idMap, cidrMap map[string]string) ([]map[string]string, error) {
	var azNames []string
	for azName := range idMap {
		azNames = append(azNames, azName)
	}
	sort.Strings(azNames)

	var azs []map[string]string
	for azIndex, azName := range azNames {
		cidr, ok := cidrMap[azName]
		if !ok {
			return []map[string]string{}, errors.New("missing AZ in terraform output: internal_az_subnet_cidr_mapping")
		}

		az, err := azify(
			azIndex+startingIndex,
			azName,
			cidr,
			idMap[azName],
		)

		if err != nil {
			return []map[string]string{}, err
		}

		azs = append(azs, az)
	}

	return azs, nil
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	ops, err := o.generateOps(state)
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

func (o OpsGenerator) generateOps(state storage.State) ([]op, error) {
	ops := []op{}
	subnets := []networkSubnet{}

	azs, err := o.availabilityZones.RetrieveAZs(state.AWS.Region)
	if err != nil {
		return []op{}, fmt.Errorf("Retrieve availability zones: %s", err)
	}

	for i := range azs {
		azOp := createOp("replace", "/azs/-", az{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: azCloudProperties{
				AvailabilityZone: fmt.Sprintf("((az%d_name))", i+1),
			},
		})
		ops = append(ops, azOp)

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

	switch state.LB.Type {
	case "cf":
		lbSecurityGroups := []map[string]string{
			{"name": "cf-router-network-properties", "lb": "((cf_router_lb_name))", "group": "((cf_router_lb_internal_security_group))"},
			{"name": "diego-ssh-proxy-network-properties", "lb": "((cf_ssh_lb_name))", "group": "((cf_ssh_lb_internal_security_group))"},
			{"name": "cf-tcp-router-network-properties", "lb": "((cf_tcp_lb_name))", "group": "((cf_tcp_lb_internal_security_group))"},
			{"name": "router-lb", "lb": "((cf_router_lb_name))", "group": "((cf_router_lb_internal_security_group))"},
			{"name": "ssh-proxy-lb", "lb": "((cf_ssh_lb_name))", "group": "((cf_ssh_lb_internal_security_group))"},
		}

		for _, details := range lbSecurityGroups {
			ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
				Name: details["name"],
				CloudProperties: lbCloudProperties{
					ELBs: []string{details["lb"]},
					SecurityGroups: []string{
						details["group"],
						"((internal_security_group))",
					},
				},
			}))
		}
	case "concourse":
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "lb",
			CloudProperties: lbCloudProperties{
				LBTargetGroups: "((concourse_lb_target_groups))",
				SecurityGroups: []string{
					"((concourse_lb_internal_security_group))",
					"((internal_security_group))",
				},
			},
		}))
	}

	return ops, nil
}

func azify(az int, azName, cidr, subnet string) (map[string]string, error) {
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
		fmt.Sprintf("az%d_name", az+1):       azName,
		fmt.Sprintf("az%d_gateway", az+1):    gateway,
		fmt.Sprintf("az%d_range", az+1):      cidr,
		fmt.Sprintf("az%d_reserved_1", az+1): fmt.Sprintf("%s-%s", firstReserved, secondReserved),
		fmt.Sprintf("az%d_reserved_2", az+1): lastReserved,
		fmt.Sprintf("az%d_static", az+1):     fmt.Sprintf("%s-%s", firstStatic, lastStatic),
		fmt.Sprintf("az%d_subnet", az+1):     subnet,
	}, nil
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
		CloudProperties: networkSubnetCloudProperties{
			Subnet:         fmt.Sprintf("((az%d_subnet))", az),
			SecurityGroups: []string{"((internal_security_group))"},
		},
	}
}
