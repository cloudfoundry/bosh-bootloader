package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type StackCreator struct {
	CreateCall struct {
		Receives struct {
			StackName string
			Template  cloudformation.Template
			Session   cloudformation.Session
		}

		Returns struct {
			Error error
		}
	}
}

func (c *StackCreator) Create(session cloudformation.Session, stackName string, template cloudformation.Template) error {
	c.CreateCall.Receives.Session = session
	c.CreateCall.Receives.StackName = stackName
	c.CreateCall.Receives.Template = template

	return c.CreateCall.Returns.Error
}
