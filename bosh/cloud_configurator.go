package bosh

import (
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
	cloudConfigInput := CloudConfigInput{
		AZs: azs,
		Subnets: []SubnetInput{
			{
				AZ:             stack.Outputs["InternalSubnet1AZ"],
				Subnet:         stack.Outputs["InternalSubnet1Name"],
				CIDR:           stack.Outputs["InternalSubnet1CIDR"],
				SecurityGroups: []string{stack.Outputs["InternalSubnet1SecurityGroup"]},
			},
			{
				AZ:             stack.Outputs["InternalSubnet2AZ"],
				Subnet:         stack.Outputs["InternalSubnet2Name"],
				CIDR:           stack.Outputs["InternalSubnet2CIDR"],
				SecurityGroups: []string{stack.Outputs["InternalSubnet2SecurityGroup"]},
			},
			{
				AZ:             stack.Outputs["InternalSubnet3AZ"],
				Subnet:         stack.Outputs["InternalSubnet3Name"],
				CIDR:           stack.Outputs["InternalSubnet3CIDR"],
				SecurityGroups: []string{stack.Outputs["InternalSubnet3SecurityGroup"]},
			},
		},
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
