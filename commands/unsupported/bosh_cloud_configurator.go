package unsupported

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
)

type CloudConfigurator struct {
	Logger    logger
	Generator cloudConfigGenerator
}

type cloudConfigGenerator interface {
	Generate(bosh.CloudConfigInput) (bosh.CloudConfig, error)
}

func NewCloudConfigurator(logger logger, generator cloudConfigGenerator) CloudConfigurator {
	return CloudConfigurator{
		Logger:    logger,
		Generator: generator,
	}
}

func (c CloudConfigurator) Configure(stack cloudformation.Stack, azs []string) error {
	cloudConfigInput := bosh.CloudConfigInput{
		AZs: azs,
		Subnets: []bosh.SubnetInput{
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

	c.Logger.Step("generating cloud config")
	cloudConfig, err := c.Generator.Generate(cloudConfigInput)
	if err != nil {
		return err
	}

	yaml, err := candiedyaml.Marshal(cloudConfig)
	if err != nil {
		return err
	}
	c.Logger.Println("cloud config:")
	c.Logger.Println(fmt.Sprintf("%s", yaml))

	return nil
}
