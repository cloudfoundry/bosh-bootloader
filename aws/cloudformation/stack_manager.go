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

type StackManager struct{}

func NewStackManager() StackManager {
	return StackManager{}
}

func (s StackManager) Create(client Session, name string, template Template) error {
	templateJson, err := json.Marshal(&template)
	if err != nil {
		return err
	}

	params := &cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		TemplateBody: aws.String(string(templateJson)),
	}

	_, err = client.CreateStack(params)
	if err != nil {
		return err
	}

	return nil
}

func (s StackManager) Update(client Session, name string, template Template) error {
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

func (s StackManager) Describe(client Session, name string) (Stack, error) {
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

func (s StackManager) CreateOrUpdate(client Session, name string, template Template) error {
	_, err := s.Describe(client, name)
	switch err {
	case StackNotFound:
		return s.Create(client, name, template)
	case nil:
		return s.Update(client, name, template)
	default:
		return err
	}
}

func (s StackManager) WaitForCompletion(client Session, name string, sleepInterval time.Duration) error {
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
		cloudformation.StackStatusUpdateComplete,
		cloudformation.StackStatusUpdateRollbackComplete,
		cloudformation.StackStatusUpdateRollbackFailed:
		return nil
	default:
		time.Sleep(sleepInterval)
		return s.WaitForCompletion(client, name, sleepInterval)
	}

	return nil
}
