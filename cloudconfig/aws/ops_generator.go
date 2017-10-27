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
	terraformManager terraformManager
}

type terraformManager interface {
	GetOutputs(storage.State) (terraform.Outputs, error)
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

func NewOpsGenerator(terraformManager terraformManager) OpsGenerator {
	return OpsGenerator{
		terraformManager: terraformManager,
	}
}

func (o OpsGenerator) GenerateVars(state storage.State) (string, error) {
	return "", nil
}

func (a OpsGenerator) Generate(state storage.State) (string, error) {
	ops, err := a.generateOps(state)
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

func (a OpsGenerator) generateOps(state storage.State) ([]op, error) {
	ops := []op{}
	subnets := []networkSubnet{}

	terraformOutputs, err := a.terraformManager.GetOutputs(state)
	if err != nil {
		return []op{}, err
	}

	internalAZSubnetIDMap := terraformOutputs.GetStringMap("internal_az_subnet_id_mapping")
	if len(internalAZSubnetIDMap) == 0 {
		return []op{}, errors.New("missing internal_az_subnet_id_mapping terraform output")
	}

	internalAZSubnetCIDRMap := terraformOutputs.GetStringMap("internal_az_subnet_cidr_mapping")
	if len(internalAZSubnetCIDRMap) == 0 {
		return []op{}, errors.New("missing internal_az_subnet_cidr_mapping terraform output")
	}

	internalSecurityGroup := terraformOutputs.GetString("internal_security_group")
	if internalSecurityGroup == "" {
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
			internalAZSubnetCIDRMap[myAZ],
			internalAZSubnetIDMap[myAZ],
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

	switch state.LB.Type {
	case "cf":
		lbSecurityGroups := []map[string]string{
			map[string]string{"name": "cf-router-network-properties", "lb": "cf_router_lb_name", "group": "cf_router_lb_internal_security_group"},
			map[string]string{"name": "diego-ssh-proxy-network-properties", "lb": "cf_ssh_lb_name", "group": "cf_ssh_lb_internal_security_group"},
			map[string]string{"name": "cf-tcp-router-network-properties", "lb": "cf_tcp_lb_name", "group": "cf_tcp_lb_internal_security_group"},
			map[string]string{"name": "router-lb", "lb": "cf_router_lb_name", "group": "cf_router_lb_internal_security_group"},
			map[string]string{"name": "ssh-proxy-lb", "lb": "cf_ssh_lb_name", "group": "cf_ssh_lb_internal_security_group"},
		}

		for _, details := range lbSecurityGroups {
			elb := terraformOutputs.GetString(details["lb"])
			if elb == "" {
				return []op{}, fmt.Errorf("missing %s terraform output", details["lb"])
			}

			grp := terraformOutputs.GetString(details["group"])
			if grp == "" {
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
	case "concourse":
		concourseLoadBalancer := terraformOutputs.GetString("concourse_lb_name")
		if concourseLoadBalancer == "" {
			return []op{}, errors.New("missing concourse_lb_name terraform output")
		}

		concourseInternalSecurityGroup := terraformOutputs.GetString("concourse_lb_internal_security_group")
		if concourseInternalSecurityGroup == "" {
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
