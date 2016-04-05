package fakes

import (
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type EC2Client struct {
	ImportKeyPairCall struct {
		Receives struct {
			Input *awsec2.ImportKeyPairInput
		}
		Returns struct {
			Error error
		}
	}

	DescribeKeyPairsCall struct {
		Receives struct {
			Input *awsec2.DescribeKeyPairsInput
		}
		Returns struct {
			Output *awsec2.DescribeKeyPairsOutput
			Error  error
		}
	}

	CreateKeyPairCall struct {
		Receives struct {
			Input *awsec2.CreateKeyPairInput
		}
		Returns struct {
			Output *awsec2.CreateKeyPairOutput
			Error  error
		}
	}

	DescribeAvailabilityZonesCall struct {
		Receives struct {
			Input *awsec2.DescribeAvailabilityZonesInput
		}
		Returns struct {
			Output *awsec2.DescribeAvailabilityZonesOutput
			Error  error
		}
	}

	DeleteKeyPairCall struct {
		Receives struct {
			Input *awsec2.DeleteKeyPairInput
		}
		Returns struct {
			Output *awsec2.DeleteKeyPairOutput
			Error  error
		}
	}

	DescribeInstancesCall struct {
		Receives struct {
			Input *awsec2.DescribeInstancesInput
		}
		Returns struct {
			Output *awsec2.DescribeInstancesOutput
			Error  error
		}
	}
}

func (c *EC2Client) ImportKeyPair(input *awsec2.ImportKeyPairInput) (*awsec2.ImportKeyPairOutput, error) {
	c.ImportKeyPairCall.Receives.Input = input

	return nil, c.ImportKeyPairCall.Returns.Error
}

func (c *EC2Client) DescribeKeyPairs(input *awsec2.DescribeKeyPairsInput) (*awsec2.DescribeKeyPairsOutput, error) {
	c.DescribeKeyPairsCall.Receives.Input = input

	return c.DescribeKeyPairsCall.Returns.Output, c.DescribeKeyPairsCall.Returns.Error
}

func (c *EC2Client) CreateKeyPair(input *awsec2.CreateKeyPairInput) (*awsec2.CreateKeyPairOutput, error) {
	c.CreateKeyPairCall.Receives.Input = input

	return c.CreateKeyPairCall.Returns.Output, c.CreateKeyPairCall.Returns.Error
}

func (c *EC2Client) DescribeAvailabilityZones(input *awsec2.DescribeAvailabilityZonesInput) (*awsec2.DescribeAvailabilityZonesOutput, error) {
	c.DescribeAvailabilityZonesCall.Receives.Input = input

	return c.DescribeAvailabilityZonesCall.Returns.Output, c.DescribeAvailabilityZonesCall.Returns.Error
}

func (c *EC2Client) DeleteKeyPair(input *awsec2.DeleteKeyPairInput) (*awsec2.DeleteKeyPairOutput, error) {
	c.DeleteKeyPairCall.Receives.Input = input

	return c.DeleteKeyPairCall.Returns.Output, c.DeleteKeyPairCall.Returns.Error
}

func (c *EC2Client) DescribeInstances(input *awsec2.DescribeInstancesInput) (*awsec2.DescribeInstancesOutput, error) {
	c.DescribeInstancesCall.Receives.Input = input

	return c.DescribeInstancesCall.Returns.Output, c.DescribeInstancesCall.Returns.Error
}
