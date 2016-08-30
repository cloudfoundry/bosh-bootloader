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
	Step(message string, a ...interface{})
	Dot()
}

type StackManager struct {
	cloudFormationClient Client
	logger               logger
}

func NewStackManager(cloudFormationClient Client, logger logger) StackManager {
	return StackManager{
		cloudFormationClient: cloudFormationClient,
		logger:               logger,
	}
}

func (s StackManager) CreateOrUpdate(name string, template templates.Template, tags Tags) error {
	_, err := s.Describe(name)
	switch err {
	case StackNotFound:
		return s.create(name, template, tags)
	case nil:
		return s.Update(name, template, tags)
	default:
		return err
	}
}

func (s StackManager) Describe(name string) (Stack, error) {
	if name == "" {
		return Stack{}, StackNotFound
	}

	output, err := s.cloudFormationClient.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})
	if err != nil {
		switch err.(type) {
		case awserr.RequestFailure:
			requestFailure := err.(awserr.RequestFailure)
			if requestFailure.StatusCode() == 400 && requestFailure.Code() == "ValidationError" &&
				requestFailure.Message() == fmt.Sprintf("Stack with id %s does not exist", name) {
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

func (s StackManager) WaitForCompletion(name string, sleepInterval time.Duration, action string) error {
	stack, err := s.Describe(name)
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
		return s.WaitForCompletion(name, sleepInterval, action)
	}

	return nil
}

func (s StackManager) Delete(name string) error {
	s.logger.Step("deleting cloudformation stack")

	_, err := s.cloudFormationClient.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: &name,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s StackManager) create(name string, template templates.Template, tags Tags) error {
	s.logger.Step("creating cloudformation stack")

	templateJson, err := json.Marshal(&template)
	if err != nil {
		return err
	}

	awsTags := tags.toAWSTags()

	params := &cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		Capabilities: []*string{aws.String("CAPABILITY_IAM"), aws.String("CAPABILITY_NAMED_IAM")},
		TemplateBody: aws.String(string(templateJson)),
		Tags:         awsTags,
	}

	_, err = s.cloudFormationClient.CreateStack(params)
	if err != nil {
		return err
	}

	return nil
}

func (s StackManager) Update(name string, template templates.Template, tags Tags) error {
	s.logger.Step("updating cloudformation stack")

	templateJson, err := json.Marshal(&template)
	if err != nil {
		return err
	}

	awsTags := tags.toAWSTags()

	params := &cloudformation.UpdateStackInput{
		StackName:    aws.String(name),
		Capabilities: []*string{aws.String("CAPABILITY_IAM"), aws.String("CAPABILITY_NAMED_IAM")},
		TemplateBody: aws.String(string(templateJson)),
		Tags:         awsTags,
	}

	_, err = s.cloudFormationClient.UpdateStack(params)
	if err != nil {
		switch err.(type) {
		case awserr.RequestFailure:
			requestFailure := err.(awserr.RequestFailure)

			if requestFailure.StatusCode() == 400 && requestFailure.Code() == "ValidationError" {
				switch requestFailure.Message() {
				case "No updates are to be performed.":
					return nil
				case fmt.Sprintf("Stack [%s] does not exist", name):
					return StackNotFound
				default:
				}
			}

			return err
		default:
			return err
		}
	}

	return nil
}

func (s StackManager) GetPhysicalIDForResource(stackName string, logicalResourceID string) (string, error) {
	describeStackResourceOutput, err := s.cloudFormationClient.DescribeStackResource(&cloudformation.DescribeStackResourceInput{
		StackName:         aws.String(stackName),
		LogicalResourceId: aws.String(logicalResourceID),
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(describeStackResourceOutput.StackResourceDetail.PhysicalResourceId), nil
}
