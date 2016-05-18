package fakes

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

type StackManager struct {
	DescribeCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Stack cloudformation.Stack
			Error error
		}
	}

	CreateOrUpdateCall struct {
		Receives struct {
			StackName string
			Template  templates.Template
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
}

func (m *StackManager) CreateOrUpdate(stackName string, template templates.Template) error {
	m.CreateOrUpdateCall.Receives.StackName = stackName
	m.CreateOrUpdateCall.Receives.Template = template

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
	m.DescribeCall.Receives.StackName = stackName

	return m.DescribeCall.Returns.Stack, m.DescribeCall.Returns.Error
}

func (m *StackManager) Delete(stackName string) error {
	m.DeleteCall.Receives.StackName = stackName

	return m.DeleteCall.Returns.Error
}
