package cloudformation

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

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

func (m InfrastructureManager) Create(keyPairName string, numberOfAvailabilityZones int, stackName string, cloudFormationClient Client) (Stack, error) {
	template := m.templateBuilder.Build(keyPairName, numberOfAvailabilityZones)

	if err := m.stackManager.CreateOrUpdate(cloudFormationClient, stackName, template); err != nil {
		return Stack{}, err
	}

	if err := m.stackManager.WaitForCompletion(cloudFormationClient, stackName, 15*time.Second); err != nil {
		return Stack{}, err
	}

	return m.stackManager.Describe(cloudFormationClient, stackName)
}

func (m InfrastructureManager) Exists(stackName string, cloudFormationClient Client) (bool, error) {
	_, err := m.stackManager.Describe(cloudFormationClient, stackName)

	switch err {
	case nil:
		return true, nil
	case StackNotFound:
		return false, nil
	default:
		return false, err
	}
}
