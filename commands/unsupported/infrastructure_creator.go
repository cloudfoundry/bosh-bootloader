package unsupported

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

type templateBuilder interface {
	Build(keypairName string) templates.Template
}

type InfrastructureCreator struct {
	templateBuilder templateBuilder
	stackManager    stackManager
}

func NewInfrastructureCreator(builder templateBuilder, stackManager stackManager) InfrastructureCreator {
	return InfrastructureCreator{
		templateBuilder: builder,
		stackManager:    stackManager,
	}
}

func (c InfrastructureCreator) Create(keyPairName string, cloudFormationClient cloudformation.Client) error {
	template := c.templateBuilder.Build(keyPairName)

	if err := c.stackManager.CreateOrUpdate(cloudFormationClient, "concourse", template); err != nil {
		return err
	}

	if err := c.stackManager.WaitForCompletion(cloudFormationClient, "concourse", 15*time.Second); err != nil {
		return err
	}

	return nil
}
