package aws

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TerraformOpsGenerator struct {
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

type isoVMExtension struct {
	Name            string
	CloudProperties securityGroupCloudProperties `yaml:"cloud_properties"`
}

type securityGroupCloudProperties struct {
	SecurityGroups []string `yaml:"security_groups"`
}

var marshal func(interface{}) ([]byte, error) = yaml.Marshal

func NewTerraformOpsGenerator(terraformManager terraformManager) TerraformOpsGenerator {
	return TerraformOpsGenerator{
		terraformManager: terraformManager,
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
	ops := []op{}
	subnets := []networkSubnet{}

	terraformOutputs, err := a.terraformManager.GetOutputs(state)
	if err != nil {
		return []op{}, err
	}

	internalAZSubnetIDMap, ok := terraformOutputs["internal_az_subnet_id_mapping"].(map[string]interface{})
	if !ok {
		return []op{}, errors.New("missing internal_az_subnet_id_mapping terraform output")
	}

	internalAZSubnetCIDRMap, ok := terraformOutputs["internal_az_subnet_cidr_mapping"].(map[string]interface{})
	if !ok {
		return []op{}, errors.New("missing internal_az_subnet_cidr_mapping terraform output")
	}

	internalSecurityGroup, ok := terraformOutputs["internal_security_group"].(string)
	if !ok {
		return []op{}, errors.New("missing internal_security_group terraform output")
	}

	var azs []string
	for myAZ, _ := range internalAZSubnetIDMap {
		azs = append(azs, myAZ)
	}
	sort.Strings(azs)

	for i, myAZ := range azs {
		azOp := createOp("replace", "/azs/-", az{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: azCloudProperties{
				AvailabilityZone: myAZ,
			},
		})
		ops = append(ops, azOp)

		subnet, err := generateNetworkSubnet(
			fmt.Sprintf("z%d", i+1),
			internalAZSubnetCIDRMap[myAZ].(string),
			internalAZSubnetIDMap[myAZ].(string),
			internalSecurityGroup,
			65,
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

	switch state.LB.Type {
	case "cf":
		sharedSGId := terraformOutputs["iso_shared_security_group_id"]

		for i := range azs {
			path := fmt.Sprintf("/networks/name=default/subnets/az=z%d/cloud_properties/security_groups/-", i+1)
			ops = append(ops, createOp("replace", path, sharedSGId))
		}

		iso1AZSubnetIDMap, ok := terraformOutputs["iso1_az_subnet_id_mapping"].(map[string]interface{})
		if !ok {
			return []op{}, errors.New("missing iso1_az_subnet_id_mapping terraform output")
		}

		iso1AZSubnetCIDRMap, ok := terraformOutputs["iso1_az_subnet_cidr_mapping"].(map[string]interface{})
		if !ok {
			return []op{}, errors.New("missing iso1_az_subnet_cidr_mapping terraform output")
		}

		iso1_subnets := []networkSubnet{}

		var iso1_azs []string
		for myAZ, _ := range iso1AZSubnetIDMap {
			iso1_azs = append(iso1_azs, myAZ)
		}
		sort.Strings(iso1_azs)

		for i, myAZ := range iso1_azs {
			azOp := createOp("replace", "/azs/-", az{
				Name: fmt.Sprintf("z%d", len(azs)+i+1),
				CloudProperties: azCloudProperties{
					AvailabilityZone: myAZ,
				},
			})
			ops = append(ops, azOp)

			subnet, err := generateNetworkSubnet(
				fmt.Sprintf("z%d", len(azs)+i+1),
				iso1AZSubnetCIDRMap[myAZ].(string),
				iso1AZSubnetIDMap[myAZ].(string),
				terraformOutputs["iso1_security_group_id"].(string),
				4,
			)
			if err != nil {
				return []op{}, err
			}

			iso1_subnets = append(iso1_subnets, subnet)
		}

		//ops = append(ops, createOp("replace", "/networks/name=default/subnets/-", iso1_subnets))
		for _, iso1_subnet := range iso1_subnets {
			ops = append(ops, createOp("replace", "/networks/name=default/subnets/-", iso1_subnet))
		}

		tfOutputs := []map[string]string{
			map[string]string{"name": "cf-router-network-properties", "lb": "cf_router_lb_name", "group": "cf_router_lb_internal_security_group"},
			map[string]string{"name": "diego-ssh-proxy-network-properties", "lb": "cf_ssh_lb_name", "group": "cf_ssh_lb_internal_security_group"},
			map[string]string{"name": "cf-tcp-router-network-properties", "lb": "cf_tcp_lb_name", "group": "cf_tcp_lb_internal_security_group"},
			map[string]string{"name": "router-lb", "lb": "cf_router_lb_name", "group": "cf_router_lb_internal_security_group"},
			map[string]string{"name": "ssh-proxy-lb", "lb": "cf_ssh_lb_name", "group": "cf_ssh_lb_internal_security_group"},
		}

		for _, details := range tfOutputs {
			elb, ok := terraformOutputs[details["lb"]].(string)
			if !ok {
				return []op{}, fmt.Errorf("missing %s terraform output", details["lb"])
			}

			grp, ok := terraformOutputs[details["group"]].(string)
			if !ok {
				return []op{}, fmt.Errorf("missing %s terraform output", details["group"])
			}

			ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
				Name: details["name"],
				CloudProperties: lbCloudProperties{
					ELBs: []string{elb},
					SecurityGroups: []string{
						grp,
						internalSecurityGroup,
					},
				},
			}))
		}

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "cf-iso1-router-network-properties",
			CloudProperties: lbCloudProperties{
				ELBs: []string{terraformOutputs["cf_iso1_router_lb_name"].(string)},
				SecurityGroups: []string{
					terraformOutputs["cf_router_lb_internal_security_group"].(string),
					terraformOutputs["iso1_security_group_id"].(string),
					internalSecurityGroup,
				},
			},
		}))

		ops = append(ops, createOp("replace", "/vm_extensions/-", isoVMExtension{
			Name: "cf-iso1-network-properties",
			CloudProperties: securityGroupCloudProperties{
				SecurityGroups: []string{
					terraformOutputs["iso1_security_group_id"].(string),
					internalSecurityGroup,
				},
			},
		}))

		//iso_subnets := terraformOutputs["iso1_az_subnet_id_mapping"].(map[string]interface{})
		//ops = append(ops, createOp("replace", "/networks/name=default/subnets/-",
		//	networkSubnet{
		//		AZ:       "z1",
		//		Gateway:  "10.0.200.1",
		//		Range:    "10.0.200.0/28",
		//		Reserved: []string{"10.0.200.2-10.0.200.3"},
		//		Static:   []string{"10.0.200.4-10.200.15"},
		//		CloudProperties: networkSubnetCloudProperties{
		//			Subnet:         iso_subnets["us-east-1a"].(string),
		//			SecurityGroups: []string{terraformOutputs["iso1_security_group_id"].(string)},
		//		},
		//	},
		//))
	case "concourse":
		concourseLoadBalancer, ok := terraformOutputs["concourse_lb_name"].(string)
		if !ok {
			return []op{}, errors.New("missing concourse_lb_name terraform output")
		}

		concourseInternalSecurityGroup, ok := terraformOutputs["concourse_lb_internal_security_group"].(string)
		if !ok {
			return []op{}, errors.New("missing concourse_lb_internal_security_group terraform output")
		}

		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "lb",
			CloudProperties: lbCloudProperties{
				ELBs: []string{concourseLoadBalancer},
				SecurityGroups: []string{
					concourseInternalSecurityGroup,
					internalSecurityGroup,
				},
			},
		}))
	}

	return ops, nil
}

func generateNetworkSubnet(az, cidr, subnet, securityGroup string, staticSize int) (networkSubnet, error) {
	parsedCidr, err := bosh.ParseCIDRBlock(cidr)
	if err != nil {
		return networkSubnet{}, err
	}

	gateway := parsedCidr.GetFirstIP().Add(1).String()
	firstReserved := parsedCidr.GetFirstIP().Add(2).String()
	secondReserved := parsedCidr.GetFirstIP().Add(3).String()
	lastReserved := parsedCidr.GetLastIP().String()
	lastStatic := parsedCidr.GetLastIP().Subtract(1).String()
	firstStatic := parsedCidr.GetLastIP().Subtract(staticSize).String()

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
