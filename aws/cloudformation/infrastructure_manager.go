package cloudformation

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
)

const bblTagKey = "bbl-env-id"

type templateBuilder interface {
	Build(keypairName string, numberOfAvailabilityZones int, lbType string, lbCertificateARN string, iamUserName string, envID string) templates.Template
}

type stackManager interface {
	CreateOrUpdate(stackName string, template templates.Template, tags Tags) error
	Update(stackName string, template templates.Template, tags Tags) error
	WaitForCompletion(stackName string, sleepInterval time.Duration, action string) error
	Describe(stackName string) (Stack, error)
	Delete(stackName string) error
	GetPhysicalIDForResource(stackName string, logicalResourceID string) (string, error)
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

func (m InfrastructureManager) Create(keyPairName string, numberOfAvailabilityZones int, stackName,
	lbType, lbCertificateARN, envID string) (Stack, error) {

	iamUserName := generateIAMUserName(envID)

	stackExists, err := m.Exists(stackName)
	if err != nil {
		return Stack{}, err
	}

	if stackExists {
		iamUserName, err = m.stackManager.GetPhysicalIDForResource(stackName, "BOSHUser")
		if err != nil {
			return Stack{}, err
		}
	}

	template := m.templateBuilder.Build(keyPairName, numberOfAvailabilityZones, lbType, lbCertificateARN, iamUserName, envID)
	tags := Tags{
		{
			Key:   bblTagKey,
			Value: envID,
		},
	}

	if err := m.stackManager.CreateOrUpdate(stackName, template, tags); err != nil {
		return Stack{}, err
	}

	if err := m.stackManager.WaitForCompletion(stackName, 15*time.Second, "applying cloudformation template"); err != nil {
		return Stack{}, err
	}

	return m.stackManager.Describe(stackName)
}

func (m InfrastructureManager) Update(keyPairName string, numberOfAvailabilityZones int, stackName, lbType,
	lbCertificateARN, envID string) (Stack, error) {

	iamUserName, err := m.stackManager.GetPhysicalIDForResource(stackName, "BOSHUser")
	if err != nil {
		return Stack{}, err
	}

	template := m.templateBuilder.Build(keyPairName, numberOfAvailabilityZones, lbType, lbCertificateARN, iamUserName, envID)

	if err := m.stackManager.Update(stackName, template, Tags{{Key: bblTagKey, Value: envID}}); err != nil {
		return Stack{}, err
	}

	if err := m.stackManager.WaitForCompletion(stackName, 15*time.Second, "applying cloudformation template"); err != nil {
		return Stack{}, err
	}

	return m.stackManager.Describe(stackName)
}

func (m InfrastructureManager) Exists(stackName string) (bool, error) {
	_, err := m.stackManager.Describe(stackName)

	switch err {
	case nil:
		return true, nil
	case StackNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (m InfrastructureManager) Describe(stackName string) (Stack, error) {
	return m.stackManager.Describe(stackName)
}

func (m InfrastructureManager) Delete(stackName string) error {
	err := m.stackManager.Delete(stackName)
	if err != nil {
		return err
	}

	err = m.stackManager.WaitForCompletion(stackName, 15*time.Second, "deleting cloudformation stack")
	if err != nil && err != StackNotFound {
		return err
	}

	return nil
}

func generateIAMUserName(envID string) string {
	return fmt.Sprintf("bosh-iam-user-%s", envID)
}
