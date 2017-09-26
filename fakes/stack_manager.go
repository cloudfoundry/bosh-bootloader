package fakes

import (
	"time"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
)

type StackManager struct {
	DescribeCall struct {
		CallCount int
		Receives  struct {
			StackName string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
		Stub func(string) (cloudformation.Stack, error)
	}

	CreateOrUpdateCall struct {
		Receives struct {
			StackName string
			Template  templates.Template
			Tags      cloudformation.Tags
		}
		Returns struct {
			Error error
		}
	}

	UpdateCall struct {
		Receives struct {
			StackName string
			Template  templates.Template
		}
		Returns struct {
			Error error
		}
	}

	WaitForCompletionCall struct {
		Receives struct {
			StackName     string
			SleepInterval time.Duration
			Action        string
		}
		Returns struct {
			Error error
		}
	}

	DeleteCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Error error
		}
	}

	GetPhysicalIDForResourceCall struct {
		Receives struct {
			StackName         string
			LogicalResourceID string
		}
		Returns struct {
			PhysicalResourceID string
			Error              error
		}
	}
}

func (m *StackManager) CreateOrUpdate(stackName string, template templates.Template, tags cloudformation.Tags) error {
	m.CreateOrUpdateCall.Receives.StackName = stackName
	m.CreateOrUpdateCall.Receives.Template = template
	m.CreateOrUpdateCall.Receives.Tags = tags

	return m.CreateOrUpdateCall.Returns.Error
}

func (m *StackManager) Update(stackName string, template templates.Template) error {
	m.UpdateCall.Receives.StackName = stackName
	m.UpdateCall.Receives.Template = template

	return m.UpdateCall.Returns.Error
}

func (m *StackManager) WaitForCompletion(stackName string, sleepInterval time.Duration, action string) error {
	m.WaitForCompletionCall.Receives.StackName = stackName
	m.WaitForCompletionCall.Receives.SleepInterval = sleepInterval
	m.WaitForCompletionCall.Receives.Action = action

	return m.WaitForCompletionCall.Returns.Error
}

func (m *StackManager) Describe(stackName string) (cloudformation.Stack, error) {
	m.DescribeCall.CallCount++
	m.DescribeCall.Receives.StackName = stackName

	if m.DescribeCall.Stub != nil {
		return m.DescribeCall.Stub(stackName)
	}

	return m.DescribeCall.Returns.Stack, m.DescribeCall.Returns.Error
}

func (m *StackManager) Delete(stackName string) error {
	m.DeleteCall.Receives.StackName = stackName

	return m.DeleteCall.Returns.Error
}

func (m *StackManager) GetPhysicalIDForResource(stackName string, logicalResourceID string) (string, error) {
	m.GetPhysicalIDForResourceCall.Receives.StackName = stackName
	m.GetPhysicalIDForResourceCall.Receives.LogicalResourceID = logicalResourceID

	return m.GetPhysicalIDForResourceCall.Returns.PhysicalResourceID, m.GetPhysicalIDForResourceCall.Returns.Error
}
