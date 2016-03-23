package cloudformation

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

const STACKNAME = "concourse"

type templateBuilder interface {
	Build(keypairName string, numberOfAvailabilityZones int) templates.Template
}

type stackManager interface {
	CreateOrUpdate(cloudFormationClient Client, stackName string, template templates.Template) error
	WaitForCompletion(cloudFormationClient Client, stackName string, sleepInterval time.Duration) error
	Describe(cloudFormationClient Client, name string) (Stack, error)
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

func (c InfrastructureCreator) Create(keyPairName string, numberOfAvailabilityZones int, cloudFormationClient Client) (Stack, error) {
	template := c.templateBuilder.Build(keyPairName, numberOfAvailabilityZones)

	if err := c.stackManager.CreateOrUpdate(cloudFormationClient, STACKNAME, template); err != nil {
		return Stack{}, err
	}

	if err := c.stackManager.WaitForCompletion(cloudFormationClient, STACKNAME, 15*time.Second); err != nil {
		return Stack{}, err
	}

	return c.stackManager.Describe(cloudFormationClient, STACKNAME)
}
