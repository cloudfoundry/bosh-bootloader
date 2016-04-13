package cloudformation

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

type templateBuilder interface {
	Build(keypairName string, numberOfAvailabilityZones int, lbType string) templates.Template
}

type stackManager interface {
	CreateOrUpdate(client Client, stackName string, template templates.Template) error
	WaitForCompletion(client Client, stackName string, sleepInterval time.Duration, action string) error
	Describe(client Client, stackName string) (Stack, error)
	Delete(client Client, stackName string) error
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

func (m InfrastructureManager) Create(keyPairName string, numberOfAvailabilityZones int, stackName string, lbType string, cloudFormationClient Client) (Stack, error) {
	template := m.templateBuilder.Build(keyPairName, numberOfAvailabilityZones, lbType)

	if err := m.stackManager.CreateOrUpdate(cloudFormationClient, stackName, template); err != nil {
		return Stack{}, err
	}

	if err := m.stackManager.WaitForCompletion(cloudFormationClient, stackName, 15*time.Second, "applying cloudformation template"); err != nil {
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

func (m InfrastructureManager) Delete(client Client, stackName string) error {
	err := m.stackManager.Delete(client, stackName)
	if err != nil {
		return err
	}

	err = m.stackManager.WaitForCompletion(client, stackName, 15*time.Second, "deleting cloudformation stack")
	if err != nil && err != StackNotFound {
		return err
	}

	return nil
}
