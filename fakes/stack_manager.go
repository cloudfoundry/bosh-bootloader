package fakes

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

type StackManager struct {
	DescribeCall struct {
		Returns struct {
			Output cloudformation.Stack
			Error  error
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

func (m *StackManager) WaitForCompletion(client cloudformation.Client, stackName string, sleepInterval time.Duration) error {
	m.WaitForCompletionCall.Receives.Client = client
	m.WaitForCompletionCall.Receives.StackName = stackName
	m.WaitForCompletionCall.Receives.SleepInterval = sleepInterval

	return m.WaitForCompletionCall.Returns.Error
}

func (s *StackManager) Describe(client cloudformation.Client, name string) (cloudformation.Stack, error) {
	return s.DescribeCall.Returns.Output, s.DescribeCall.Returns.Error
}
