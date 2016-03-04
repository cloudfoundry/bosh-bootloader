package cloudformation

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var StackNotFound error = errors.New("stack not found")

type StackManager struct {
	logger logger
}

func NewStackManager(logger logger) StackManager {
	return StackManager{
		logger: logger,
	}
}

func (s StackManager) CreateOrUpdate(client Client, name string, template Template) error {
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

			return Stack{
				Name:   *s.StackName,
				Status: status,
			}, nil
		}
	}

	return Stack{}, StackNotFound
}

func (s StackManager) WaitForCompletion(client Client, name string, sleepInterval time.Duration) error {
	output, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})
	if err != nil {
		return err
	}

	status := "UNKNOWN"
	for _, s := range output.Stacks {
		if s.StackName != nil && *s.StackName == name {
			if s.StackStatus != nil {
				status = *s.StackStatus
			}

			break
		}
	}

	switch status {
	case cloudformation.StackStatusCreateComplete,
		cloudformation.StackStatusCreateFailed,
		cloudformation.StackStatusRollbackComplete,
		cloudformation.StackStatusRollbackFailed,
		cloudformation.StackStatusUpdateComplete,
		cloudformation.StackStatusUpdateRollbackComplete,
		cloudformation.StackStatusUpdateRollbackFailed:
		s.logger.Step("finished applying cloudformation template")
		return nil
	default:
		s.logger.Dot()
		time.Sleep(sleepInterval)
		return s.WaitForCompletion(client, name, sleepInterval)
	}

	return nil
}

func (s StackManager) create(client Client, name string, template Template) error {
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

func (s StackManager) update(client Client, name string, template Template) error {
	s.logger.Step("updating cloudformation stack")

	templateJson, err := json.Marshal(&template)
	if err != nil {
		return err
	}

	params := &cloudformation.UpdateStackInput{
		StackName:    aws.String(name),
		TemplateBody: aws.String(string(templateJson)),
	}

	_, err = client.UpdateStack(params)
	if err != nil {
		if err != nil {
			switch err.(type) {
			case awserr.RequestFailure:
				if err.(awserr.RequestFailure).StatusCode() == 400 {
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
