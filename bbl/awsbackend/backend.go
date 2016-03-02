package awsbackend

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rosenhouse/awsfaker"
)

type Backend struct {
	KeyPairs *KeyPairs
	Stacks   *Stacks
}

func New() *Backend {
	return &Backend{
		KeyPairs: NewKeyPairs(),
		Stacks:   NewStacks(),
	}
}

func (b *Backend) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	keyPair := KeyPair{
		Name: *input.KeyName,
	}

	b.KeyPairs.Set(keyPair)

	return &ec2.CreateKeyPairOutput{
		KeyName: aws.String(keyPair.Name),
	}, nil
}

func (b *Backend) DescribeKeyPairs(input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
	var keyPairs []KeyPair
	for _, name := range input.KeyNames {
		keyPair, ok := b.KeyPairs.Get(*name)
		if !ok {
			return nil, &awsfaker.ErrorResponse{
				HTTPStatusCode:  http.StatusBadRequest,
				AWSErrorCode:    "InvalidKeyPair.NotFound",
				AWSErrorMessage: fmt.Sprintf("The key pair '%s' does not exist", name),
			}
		}
		keyPairs = append(keyPairs, keyPair)
	}

	var keyPairInfos []*ec2.KeyPairInfo
	for _, keyPair := range keyPairs {
		keyPairInfos = append(keyPairInfos, &ec2.KeyPairInfo{
			KeyName: aws.String(keyPair.Name),
		})
	}

	return &ec2.DescribeKeyPairsOutput{
		KeyPairs: keyPairInfos,
	}, nil
}

func (b *Backend) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	stack := Stack{
		Name: *input.StackName,
	}
	b.Stacks.Set(stack)

	return &cloudformation.CreateStackOutput{}, nil
}

func (b *Backend) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	name := *input.StackName
	stack, ok := b.Stacks.Get(name)
	if !ok {
		return nil, &awsfaker.ErrorResponse{
			HTTPStatusCode:  http.StatusBadRequest,
			AWSErrorCode:    "ValidationError",
			AWSErrorMessage: fmt.Sprintf("Stack [%s] does not exist", name),
		}
	}

	stack.WasUpdated = true
	b.Stacks.Set(stack)

	return &cloudformation.UpdateStackOutput{}, nil
}

func (b *Backend) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	name := *input.StackName
	stack, ok := b.Stacks.Get(name)
	if !ok {
		return nil, &awsfaker.ErrorResponse{
			HTTPStatusCode:  http.StatusBadRequest,
			AWSErrorCode:    "ValidationError",
			AWSErrorMessage: fmt.Sprintf("Stack with id %s does not exist", name),
		}
	}

	return &cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackName:   aws.String(stack.Name),
				StackStatus: aws.String("CREATE_COMPLETE"),
			},
		},
	}, nil
}
