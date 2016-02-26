package fakes

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
)

type StackManager struct {
	CreateOrUpdateCall struct {
		Receives struct {
			StackName string
			Template  cloudformation.Template
			Session   cloudformation.Session
		}
		Returns struct {
			Error error
		}
	}

	WaitForCompletionCall struct {
		Receives struct {
			Session       cloudformation.Session
			StackName     string
			SleepInterval time.Duration
		}
		Returns struct {
			Error error
		}
	}
}

func (m *StackManager) CreateOrUpdate(session cloudformation.Session, stackName string, template cloudformation.Template) error {
	m.CreateOrUpdateCall.Receives.Session = session
	m.CreateOrUpdateCall.Receives.StackName = stackName
	m.CreateOrUpdateCall.Receives.Template = template

	return m.CreateOrUpdateCall.Returns.Error
}

func (m *StackManager) WaitForCompletion(session cloudformation.Session, stackName string, sleepInterval time.Duration) error {
	m.WaitForCompletionCall.Receives.Session = session
	m.WaitForCompletionCall.Receives.StackName = stackName
	m.WaitForCompletionCall.Receives.SleepInterval = sleepInterval

	return m.WaitForCompletionCall.Returns.Error
}
