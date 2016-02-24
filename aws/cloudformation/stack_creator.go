package cloudformation

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type StackCreator struct{}

func NewStackCreator() StackCreator {
	return StackCreator{}
}

func (s StackCreator) Create(cloudFormationClient Session, stackName string, template Template) error {
	templateJson, err := json.Marshal(&template)
	if err != nil {
		return err
	}

	params := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(string(templateJson)),
	}

	_, err = cloudFormationClient.CreateStack(params)
	if err != nil {
		return err
	}

	return nil
}
