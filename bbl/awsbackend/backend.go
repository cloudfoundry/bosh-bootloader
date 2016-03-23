package awsbackend

import (
	"fmt"
	"net/http"
	"reflect"

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
			KeyName:        aws.String(keyPair.Name),
			KeyFingerprint: aws.String("some-fingerprint"),
		})
	}

	return &ec2.DescribeKeyPairsOutput{
		KeyPairs: keyPairInfos,
	}, nil
}

func (b *Backend) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	stack := Stack{
		Name:     *input.StackName,
		Template: *input.TemplateBody,
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

func (b *Backend) DescribeAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error) {
	validInput := &ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("region-name"),
			Values: []*string{aws.String("some-region")},
		}},
	}

	if !reflect.DeepEqual(input, validInput) {
		return nil, nil
	}

	return &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []*ec2.AvailabilityZone{
			{ZoneName: aws.String("us-east-1a")},
			{ZoneName: aws.String("us-east-1b")},
			{ZoneName: aws.String("us-east-1c")},
			{ZoneName: aws.String("us-east-1e")},
		},
	}, nil
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
				Outputs: []*cloudformation.Output{
					{
						OutputKey:   aws.String("BOSHEIP"),
						OutputValue: aws.String("192.168.1.1"),
					},
					{
						OutputKey:   aws.String("InternalSubnet1CIDR"),
						OutputValue: aws.String("10.0.16.0/20"),
					},
					{
						OutputKey:   aws.String("InternalSubnet2CIDR"),
						OutputValue: aws.String("10.0.32.0/20"),
					},
					{
						OutputKey:   aws.String("InternalSubnet3CIDR"),
						OutputValue: aws.String("10.0.48.0/20"),
					},
					{
						OutputKey:   aws.String("InternalSubnet1AZ"),
						OutputValue: aws.String("us-east-1a"),
					},
					{
						OutputKey:   aws.String("InternalSubnet2AZ"),
						OutputValue: aws.String("us-east-1b"),
					},
					{
						OutputKey:   aws.String("InternalSubnet3AZ"),
						OutputValue: aws.String("us-east-1c"),
					},
					{
						OutputKey:   aws.String("InternalSubnet1Name"),
						OutputValue: aws.String("some-subnet-1"),
					},
					{
						OutputKey:   aws.String("InternalSubnet2Name"),
						OutputValue: aws.String("some-subnet-2"),
					},
					{
						OutputKey:   aws.String("InternalSubnet3Name"),
						OutputValue: aws.String("some-subnet-3"),
					},
					{
						OutputKey:   aws.String("InternalSubnet1SecurityGroup"),
						OutputValue: aws.String("some-security-group-1"),
					},
					{
						OutputKey:   aws.String("InternalSubnet2SecurityGroup"),
						OutputValue: aws.String("some-security-group-2"),
					},
					{
						OutputKey:   aws.String("InternalSubnet3SecurityGroup"),
						OutputValue: aws.String("some-security-group-3"),
					},
				},
			},
		},
	}, nil
}
