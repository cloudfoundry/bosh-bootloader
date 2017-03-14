package aws

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGenerator struct {
	availabilityZoneRetriever availabilityZoneRetriever
	infrastructureManager     infrastructureManager
}

type availabilityZoneRetriever interface {
	Retrieve(string) ([]string, error)
}

type infrastructureManager interface {
	Describe(stackName string) (cloudformation.Stack, error)
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

func NewOpsGenerator(availabilityZoneRetriever availabilityZoneRetriever, infrastructureManager infrastructureManager) OpsGenerator {
	return OpsGenerator{
		availabilityZoneRetriever: availabilityZoneRetriever,
		infrastructureManager:     infrastructureManager,
	}
}

func (a OpsGenerator) Generate(state storage.State) (string, error) {
	ops, err := a.generateAWSOps(state)
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

func (a OpsGenerator) generateAWSOps(state storage.State) ([]op, error) {
	azs, err := a.availabilityZoneRetriever.Retrieve(state.AWS.Region)
	if err != nil {
		return []op{}, err
	}

	ops := []op{}
	for i, awsAZ := range azs {
		op := createOp("replace", "/azs/-", az{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: azCloudProperties{
				AvailabilityZone: awsAZ,
			},
		})
		ops = append(ops, op)
	}

	stack, err := a.infrastructureManager.Describe(state.Stack.Name)
	if err != nil {
		return []op{}, err
	}

	subnets := []networkSubnet{}
	for i := range azs {
		subnet, err := generateNetworkSubnet(
			fmt.Sprintf("z%d", i+1),
			stack.Outputs[fmt.Sprintf("InternalSubnet%dCIDR", i+1)],
			stack.Outputs[fmt.Sprintf("InternalSubnet%dName", i+1)],
			stack.Outputs["InternalSecurityGroup"],
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

	if value := stack.Outputs["CFRouterLoadBalancer"]; value != "" {
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "router-lb",
			CloudProperties: lbCloudProperties{
				ELBs: []string{value},
				SecurityGroups: []string{
					stack.Outputs["CFRouterInternalSecurityGroup"],
					stack.Outputs["InternalSecurityGroup"],
				},
			},
		}))
	}

	if value := stack.Outputs["CFSSHProxyLoadBalancer"]; value != "" {
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "ssh-proxy-lb",
			CloudProperties: lbCloudProperties{
				ELBs: []string{value},
				SecurityGroups: []string{
					stack.Outputs["CFSSHProxyInternalSecurityGroup"],
					stack.Outputs["InternalSecurityGroup"],
				},
			},
		}))
	}

	if value := stack.Outputs["ConcourseLoadBalancer"]; value != "" {
		ops = append(ops, createOp("replace", "/vm_extensions/-", lb{
			Name: "lb",
			CloudProperties: lbCloudProperties{
				ELBs: []string{value},
				SecurityGroups: []string{
					stack.Outputs["ConcourseInternalSecurityGroup"],
					stack.Outputs["InternalSecurityGroup"],
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
