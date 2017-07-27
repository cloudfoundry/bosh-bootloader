package aws

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type availabilityZoneRetriever interface {
	Retrieve(string) ([]string, error)
}

type CloudFormationOpsGenerator struct {
	availabilityZoneRetriever availabilityZoneRetriever
	infrastructureManager     infrastructureManager
}

type infrastructureManager interface {
	Describe(stackName string) (cloudformation.Stack, error)
}

func NewCloudFormationOpsGenerator(availabilityZoneRetriever availabilityZoneRetriever, infrastructureManager infrastructureManager) CloudFormationOpsGenerator {
	return CloudFormationOpsGenerator{
		availabilityZoneRetriever: availabilityZoneRetriever,
		infrastructureManager:     infrastructureManager,
	}
}

func (a CloudFormationOpsGenerator) Generate(state storage.State) (string, error) {
	ops, err := a.generateCloudFormationAWSOps(state)
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

func (a CloudFormationOpsGenerator) generateCloudFormationAWSOps(state storage.State) ([]op, error) {
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
