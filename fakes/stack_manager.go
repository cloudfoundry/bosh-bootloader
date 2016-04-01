package fakes

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

type StackManager struct {
	DescribeCall struct {
		Receives struct {
			Client    cloudformation.Client
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
			Client    cloudformation.Client
		}
		Returns struct {
			Error error
		}
	}

	WaitForCompletionCall struct {
		Receives struct {
			Client        cloudformation.Client
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
			Client    cloudformation.Client
			StackName string
		}
		Returns struct {
			Error error
		}
	}
}

func (m *StackManager) CreateOrUpdate(client cloudformation.Client, stackName string, template templates.Template) error {
	m.CreateOrUpdateCall.Receives.Client = client
	m.CreateOrUpdateCall.Receives.StackName = stackName
	m.CreateOrUpdateCall.Receives.Template = template

	return m.CreateOrUpdateCall.Returns.Error
}

func (m *StackManager) WaitForCompletion(client cloudformation.Client, stackName string, sleepInterval time.Duration, action string) error {
	m.WaitForCompletionCall.Receives.Client = client
	m.WaitForCompletionCall.Receives.StackName = stackName
	m.WaitForCompletionCall.Receives.SleepInterval = sleepInterval
	m.WaitForCompletionCall.Receives.Action = action

	return m.WaitForCompletionCall.Returns.Error
}

func (m *StackManager) Describe(client cloudformation.Client, stackName string) (cloudformation.Stack, error) {
	m.DescribeCall.Receives.Client = client
	m.DescribeCall.Receives.StackName = stackName

	return m.DescribeCall.Returns.Stack, m.DescribeCall.Returns.Error
}

func (m *StackManager) Delete(client cloudformation.Client, stackName string) error {
	m.DeleteCall.Receives.Client = client
	m.DeleteCall.Receives.StackName = stackName

	return m.DeleteCall.Returns.Error
}
