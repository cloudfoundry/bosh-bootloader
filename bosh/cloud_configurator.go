package bosh

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"
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

type cloudConfigGenerator interface {
	Generate(CloudConfigInput) (CloudConfig, error)
}

func NewCloudConfigurator(logger logger, generator cloudConfigGenerator) CloudConfigurator {
	return CloudConfigurator{
		logger:    logger,
		generator: generator,
	}
}

func (c CloudConfigurator) Configure(stack cloudformation.Stack, azs []string, boshClient Client) error {
	var subnets []SubnetInput
	for az := range azs {
		az++
		subnets = append(subnets, SubnetInput{
			AZ:             stack.Outputs[fmt.Sprintf("InternalSubnet%dAZ", az)],
			Subnet:         stack.Outputs[fmt.Sprintf("InternalSubnet%dName", az)],
			CIDR:           stack.Outputs[fmt.Sprintf("InternalSubnet%dCIDR", az)],
			SecurityGroups: []string{stack.Outputs[fmt.Sprintf("InternalSubnet%dSecurityGroup", az)]},
		})
	}

	cloudConfigInput := CloudConfigInput{
		AZs:     azs,
		Subnets: subnets,
		LBs:     c.populateLBs(stack),
	}

	c.logger.Step("generating cloud config")
	cloudConfig, err := c.generator.Generate(cloudConfigInput)
	if err != nil {
		return err
	}

	yaml, err := candiedyaml.Marshal(cloudConfig)
	if err != nil {
		return err
	}

	c.logger.Step("applying cloud config")
	if err := boshClient.UpdateCloudConfig(yaml); err != nil {
		return err
	}

	return nil
}

func (CloudConfigurator) populateLBs(stack cloudformation.Stack) map[string]string {
	lbs := map[string]string{}
	if stack.Outputs["ConcourseLoadBalancer"] != "" {
		lbs["lb"] = stack.Outputs["ConcourseLoadBalancer"]
	}

	if stack.Outputs["CFLB"] != "" {
		lbs["cf-lb"] = stack.Outputs["CFLB"]
	}
	return lbs
}
