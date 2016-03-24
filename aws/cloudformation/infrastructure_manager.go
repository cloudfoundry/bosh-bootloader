package cloudformation

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

const STACK_NAME = "concourse"

type templateBuilder interface {
	Build(keypairName string, numberOfAvailabilityZones int) templates.Template
}

type stackManager interface {
	CreateOrUpdate(cloudFormationClient Client, stackName string, template templates.Template) error
	WaitForCompletion(cloudFormationClient Client, stackName string, sleepInterval time.Duration) error
	Describe(cloudFormationClient Client, name string) (Stack, error)
}

type InfrastructureManager struct {
	templateBuilder templateBuilder
	stackManager    stackManager
}

func NewInfrastructureManager(builder templateBuilder, stackManager stackManager) InfrastructureManager {
	return InfrastructureManager{
		templateBuilder: builder,
		stackManager:    stackManager,
	}
}

func (m InfrastructureManager) Create(keyPairName string, numberOfAvailabilityZones int, cloudFormationClient Client) (Stack, error) {
	template := m.templateBuilder.Build(keyPairName, numberOfAvailabilityZones)

	if err := m.stackManager.CreateOrUpdate(cloudFormationClient, STACK_NAME, template); err != nil {
		return Stack{}, err
	}

	if err := m.stackManager.WaitForCompletion(cloudFormationClient, STACK_NAME, 15*time.Second); err != nil {
		return Stack{}, err
	}

	return m.stackManager.Describe(cloudFormationClient, STACK_NAME)
}

func (m InfrastructureManager) Exists(cloudFormationClient Client) (bool, error) {
	_, err := m.stackManager.Describe(cloudFormationClient, STACK_NAME)

	switch err {
	case nil:
		return true, nil
	case StackNotFound:
		return false, nil
	default:
		return false, err
	}
}
