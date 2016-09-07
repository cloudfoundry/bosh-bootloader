package awsbackend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/testhelpers"
	"github.com/rosenhouse/awsfaker"
)

type Backend struct {
	KeyPairs        *KeyPairs
	Stacks          *Stacks
	LoadBalancers   *LoadBalancers
	Instances       *Instances
	Certificates    *Certificates
	boshDirectorURL string

	CreateKeyPairCallCount int64
	CreateStackCallCount   int64
}

func New(boshDirectorURL string) *Backend {
	return &Backend{
		KeyPairs:        NewKeyPairs(),
		Stacks:          NewStacks(),
		Instances:       NewInstances(),
		boshDirectorURL: boshDirectorURL,
		LoadBalancers:   NewLoadBalancers(),
		Certificates:    NewCertificates(),
	}
}

func (b *Backend) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	keyPair := KeyPair{
		Name: *input.KeyName,
	}
	atomic.AddInt64(&b.CreateKeyPairCallCount, 1)
	b.KeyPairs.Set(keyPair)

	if err := b.KeyPairs.CreateKeyPairReturnError(); err != nil {
		return nil, err
	}

	return &ec2.CreateKeyPairOutput{
		KeyName:     aws.String(keyPair.Name),
		KeyMaterial: aws.String(testhelpers.PRIVATE_KEY),
	}, nil
}

func (b *Backend) DeleteKeyPair(input *ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error) {
	if err := b.KeyPairs.DeleteKeyPairReturnError(); err != nil {
		return nil, err
	}

	b.KeyPairs.Delete(*input.KeyName)

	return &ec2.DeleteKeyPairOutput{}, nil
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

func (b *Backend) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	reservations := []*ec2.Reservation{}
	for _, instance := range b.Instances.Get() {
		if aws.StringValue(input.Filters[0].Name) == "vpc-id" &&
			aws.StringValue(input.Filters[0].Values[0]) == instance.VPCID {

			reservations = append(reservations, &ec2.Reservation{
				Instances: []*ec2.Instance{
					{
						Tags: []*ec2.Tag{{
							Key:   aws.String("Name"),
							Value: aws.String(instance.Name),
						}},
					},
				},
			})
		}
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: reservations,
	}, nil
}

func (b *Backend) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	stack := Stack{
		Name:     *input.StackName,
		Template: *input.TemplateBody,
	}
	atomic.AddInt64(&b.CreateStackCallCount, 1)
	b.Stacks.Set(stack)

	if err := b.Stacks.CreateStackReturnError(); err != nil {
		return nil, err
	}

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
	stack.Template = *input.TemplateBody
	b.Stacks.Set(stack)

	return &cloudformation.UpdateStackOutput{}, nil
}

func (b *Backend) DeleteStack(input *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	if err := b.Stacks.DeleteStackReturnError(); err != nil {
		return nil, err
	}

	name := *input.StackName
	b.Stacks.Delete(name)

	return &cloudformation.DeleteStackOutput{}, nil
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
		},
	}, nil
}

func (b *Backend) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	var loadBalancer LoadBalancer

	if len(input.LoadBalancerNames) > 0 {
		loadBalancer, _ = b.LoadBalancers.Get(aws.StringValue(input.LoadBalancerNames[0]))
	}

	output := &elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
			{
				Instances: []*elb.Instance{},
			},
		},
	}

	for _, name := range loadBalancer.Instances {
		instance := &elb.Instance{
			InstanceId: aws.String(name),
		}

		output.LoadBalancerDescriptions[0].Instances = append(output.LoadBalancerDescriptions[0].Instances, instance)
	}

	return output, nil
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

	stackOutput := &cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackName:   aws.String(stack.Name),
				StackStatus: aws.String("CREATE_COMPLETE"),
				Outputs: []*cloudformation.Output{
					{
						OutputKey:   aws.String("BOSHEIP"),
						OutputValue: aws.String("127.0.0.1"),
					},
					{
						OutputKey:   aws.String("BOSHURL"),
						OutputValue: aws.String(b.boshDirectorURL),
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
						OutputKey:   aws.String("InternalSecurityGroup"),
						OutputValue: aws.String("some-internal-security-group"),
					},
					{
						OutputKey:   aws.String("VPCID"),
						OutputValue: aws.String("some-vpc-id"),
					},
				},
			},
		},
	}

	if stack.Template != "" {
		var template templates.Template
		err := json.Unmarshal([]byte(stack.Template), &template)
		if err != nil {
			return nil, err
		}

		if _, ok := template.Resources["ConcourseLoadBalancer"]; ok {
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("ConcourseLoadBalancer"),
				OutputValue: aws.String("some-concourse-lb"),
			})
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("ConcourseInternalSecurityGroup"),
				OutputValue: aws.String("some-concourse-internal-security-group"),
			})
		}

		if _, ok := template.Resources["CFRouterLoadBalancer"]; ok {
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("CFRouterLoadBalancer"),
				OutputValue: aws.String("some-cf-router-lb"),
			})
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("CFRouterLoadBalancerURL"),
				OutputValue: aws.String("some-cf-router-lb-url"),
			})
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("CFSSHProxyLoadBalancer"),
				OutputValue: aws.String("some-cf-ssh-proxy-lb"),
			})
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("CFSSHProxyLoadBalancerURL"),
				OutputValue: aws.String("some-cf-ssh-proxy-lb-url"),
			})
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("CFRouterInternalSecurityGroup"),
				OutputValue: aws.String("some-cf-router-internal-security-group"),
			})
			stackOutput.Stacks[0].Outputs = append(stackOutput.Stacks[0].Outputs, &cloudformation.Output{
				OutputKey:   aws.String("CFSSHProxyInternalSecurityGroup"),
				OutputValue: aws.String("some-cf-ssh-proxy-internal-security-group"),
			})
		}
	}

	return stackOutput, nil
}

func (b *Backend) DescribeStackResource(input *cloudformation.DescribeStackResourceInput) (*cloudformation.DescribeStackResourceOutput, error) {
	return &cloudformation.DescribeStackResourceOutput{
		StackResourceDetail: &cloudformation.StackResourceDetail{
			ResourceType:       aws.String("AWS::IAM::User"),
			StackName:          aws.String("some-stack-name"),
			PhysicalResourceId: aws.String("some-stack-name-BOSHUser-random"),
			LogicalResourceId:  aws.String("BOSHUser"),
		},
	}, nil
}

func (b *Backend) GetServerCertificate(input *iam.GetServerCertificateInput) (*iam.GetServerCertificateOutput, error) {
	certificateName := aws.StringValue(input.ServerCertificateName)

	if certificate, ok := b.Certificates.Get(certificateName); ok {
		return &iam.GetServerCertificateOutput{
			ServerCertificate: &iam.ServerCertificate{
				CertificateBody: aws.String(certificate.CertificateBody),
				ServerCertificateMetadata: &iam.ServerCertificateMetadata{
					Path:                  aws.String("some-certificate-path"),
					Arn:                   aws.String("some-certificate-arn"),
					ServerCertificateId:   aws.String("some-server-certificate-id"),
					ServerCertificateName: input.ServerCertificateName,
				},
			},
		}, nil
	}

	return nil, &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusNotFound,
		AWSErrorCode:    "NoSuchEntity",
		AWSErrorMessage: fmt.Sprintf("The Server Certificate with name %s cannot be found.", certificateName),
	}
}

func (b *Backend) UploadServerCertificate(input *iam.UploadServerCertificateInput) (*iam.UploadServerCertificateOutput, error) {
	certificateName := aws.StringValue(input.ServerCertificateName)

	if _, ok := b.Certificates.Get(certificateName); !ok {
		b.Certificates.Set(Certificate{
			Name:            certificateName,
			CertificateBody: aws.StringValue(input.CertificateBody),
			PrivateKey:      aws.StringValue(input.PrivateKey),
			Chain:           aws.StringValue(input.CertificateChain),
		})
		return nil, nil
	}

	return nil, &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusConflict,
		AWSErrorCode:    "EntityAlreadyExists",
		AWSErrorMessage: fmt.Sprintf("The Server Certificate with name %s already exists.", certificateName),
	}
}

func (b *Backend) DeleteServerCertificate(input *iam.DeleteServerCertificateInput) (*iam.DeleteServerCertificateOutput, error) {
	certificateName := aws.StringValue(input.ServerCertificateName)

	if _, ok := b.Certificates.Get(certificateName); ok {
		b.Certificates.Delete(certificateName)
		return &iam.DeleteServerCertificateOutput{}, nil
	}

	return nil, &awsfaker.ErrorResponse{
		HTTPStatusCode:  http.StatusNotFound,
		AWSErrorCode:    "NoSuchEntity",
		AWSErrorMessage: fmt.Sprintf("The Server Certificate with name %s cannot be found.", certificateName),
	}
}
