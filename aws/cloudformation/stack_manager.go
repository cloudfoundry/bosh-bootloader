package cloudformation

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

var StackNotFound error = errors.New("stack not found")

type logger interface {
	Step(message string)
	Dot()
}

type StackManager struct {
	logger logger
}

func NewStackManager(logger logger) StackManager {
	return StackManager{
		logger: logger,
	}
}

func (s StackManager) CreateOrUpdate(client Client, name string, template templates.Template) error {
	_, err := s.Describe(client, name)
	switch err {
	case StackNotFound:
		return s.create(client, name, template)
	case nil:
		return s.update(client, name, template)
	default:
		return err
	}
}

func (s StackManager) Describe(client Client, name string) (Stack, error) {
	if name == "" {
		return Stack{}, StackNotFound
	}

	output, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})
	if err != nil {
		switch err.(type) {
		case awserr.RequestFailure:
			if err.(awserr.RequestFailure).StatusCode() == 400 {
				return Stack{}, StackNotFound
			}

			return Stack{}, err
		default:
			return Stack{}, err
		}
	}

	for _, s := range output.Stacks {
		if s.StackName != nil && *s.StackName == name {
			status := "UNKNOWN"

			if s.StackStatus != nil {
				status = *s.StackStatus
			}

			stack := Stack{
				Name:    *s.StackName,
				Status:  status,
				Outputs: map[string]string{},
			}

			for _, output := range s.Outputs {
				if output.OutputKey == nil {
					return Stack{}, errors.New("failed to parse outputs")
				}

				value := ""
				if output.OutputValue != nil {
					value = *output.OutputValue
				}

				stack.Outputs[*output.OutputKey] = value
			}

			return stack, nil
		}
	}

	return Stack{}, StackNotFound
}

func (s StackManager) WaitForCompletion(client Client, name string, sleepInterval time.Duration, action string) error {
	stack, err := s.Describe(client, name)
	if err != nil {
		if err == StackNotFound {
			s.logger.Step(fmt.Sprintf("finished %s", action))
			return nil
		}

		return err
	}

	switch stack.Status {
	case cloudformation.StackStatusCreateComplete,
		cloudformation.StackStatusUpdateComplete,
		cloudformation.StackStatusDeleteComplete:
		s.logger.Step(fmt.Sprintf("finished %s", action))
		return nil
	case cloudformation.StackStatusCreateFailed,
		cloudformation.StackStatusRollbackComplete,
		cloudformation.StackStatusRollbackFailed,
		cloudformation.StackStatusUpdateRollbackComplete,
		cloudformation.StackStatusUpdateRollbackFailed,
		cloudformation.StackStatusDeleteFailed:
		return fmt.Errorf(`CloudFormation failure on stack '%s'.
Check the AWS console for error events related to this stack,
and/or open a GitHub issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues.`, name)
	default:
		s.logger.Dot()
		time.Sleep(sleepInterval)
		return s.WaitForCompletion(client, name, sleepInterval, action)
	}

	return nil
}

func (s StackManager) Delete(client Client, name string) error {
	s.logger.Step("deleting cloudformation stack")

	_, err := client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: &name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s StackManager) create(client Client, name string, template templates.Template) error {
	s.logger.Step("creating cloudformation stack")

	templateJson, err := json.Marshal(&template)
	if err != nil {
		return err
	}

	params := &cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		TemplateBody: aws.String(string(templateJson)),
	}

	_, err = client.CreateStack(params)
	if err != nil {
		return err
	}

	return nil
}

func (s StackManager) update(client Client, name string, template templates.Template) error {
	s.logger.Step("updating cloudformation stack")

	templateJson, err := json.Marshal(&template)
	if err != nil {
		return err
	}

	params := &cloudformation.UpdateStackInput{
		StackName:    aws.String(name),
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		TemplateBody: aws.String(string(templateJson)),
	}

	_, err = client.UpdateStack(params)
	if err != nil {
		if err != nil {
			switch err.(type) {
			case awserr.RequestFailure:
				requestFailure := err.(awserr.RequestFailure)
				if requestFailure.StatusCode() == 400 && requestFailure.Code() == "ValidationError" &&
					requestFailure.Message() == "No updates are to be performed." {
					return nil
				}
				return err
			default:
				return err
			}
		}
	}

	return nil
}
