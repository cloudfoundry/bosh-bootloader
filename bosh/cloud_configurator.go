package bosh

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
)

type CloudConfigurator struct {
	logger    logger
	generator cloudConfigGenerator
}

type logger interface {
	Step(message string)
	Println(string)
}

func NewCloudConfigurator(logger logger, generator cloudConfigGenerator) CloudConfigurator {
	return CloudConfigurator{
		logger:    logger,
		generator: generator,
	}
}

func (c CloudConfigurator) Configure(stack cloudformation.Stack, azs []string) CloudConfigInput {
	var subnets []SubnetInput
	for az := range azs {
		az++
		subnets = append(subnets, SubnetInput{
			AZ:             stack.Outputs[fmt.Sprintf("InternalSubnet%dAZ", az)],
			Subnet:         stack.Outputs[fmt.Sprintf("InternalSubnet%dName", az)],
			CIDR:           stack.Outputs[fmt.Sprintf("InternalSubnet%dCIDR", az)],
			SecurityGroups: []string{stack.Outputs["InternalSecurityGroup"]},
		})
	}

	cloudConfigInput := CloudConfigInput{
		AZs:     azs,
		Subnets: subnets,
		LBs:     c.populateLBs(stack),
	}

	return cloudConfigInput
}

func (CloudConfigurator) populateLBs(stack cloudformation.Stack) []LoadBalancerExtension {
	lbs := []LoadBalancerExtension{}

	if value := stack.Outputs["ConcourseLoadBalancer"]; value != "" {
		lbs = append(lbs, LoadBalancerExtension{
			Name:    "lb",
			ELBName: value,
			SecurityGroups: []string{
				stack.Outputs["ConcourseInternalSecurityGroup"],
				stack.Outputs["InternalSecurityGroup"],
			},
		})
	}

	if value := stack.Outputs["CFRouterLoadBalancer"]; value != "" {
		lbs = append(lbs, LoadBalancerExtension{
			Name:    "router-lb",
			ELBName: value,
			SecurityGroups: []string{
				stack.Outputs["CFRouterInternalSecurityGroup"],
				stack.Outputs["InternalSecurityGroup"],
			},
		})
	}

	if value := stack.Outputs["CFSSHProxyLoadBalancer"]; value != "" {
		lbs = append(lbs, LoadBalancerExtension{
			Name:    "ssh-proxy-lb",
			ELBName: value,
			SecurityGroups: []string{
				stack.Outputs["CFSSHProxyInternalSecurityGroup"],
				stack.Outputs["InternalSecurityGroup"],
			},
		})
	}

	return lbs
}
