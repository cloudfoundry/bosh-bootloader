package fakes

import (
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type AWSEC2Client struct {
	DescribeAvailabilityZonesCall struct {
		Receives struct {
			Input *awsec2.DescribeAvailabilityZonesInput
		}
		Returns struct {
			Output *awsec2.DescribeAvailabilityZonesOutput
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

	DescribeVpcsCall struct {
		Receives struct {
			Input *awsec2.DescribeVpcsInput
		}
		Returns struct {
			Output *awsec2.DescribeVpcsOutput
			Error  error
		}
	}
}

func (c *AWSEC2Client) DescribeAvailabilityZones(input *awsec2.DescribeAvailabilityZonesInput) (*awsec2.DescribeAvailabilityZonesOutput, error) {
	c.DescribeAvailabilityZonesCall.Receives.Input = input

	return c.DescribeAvailabilityZonesCall.Returns.Output, c.DescribeAvailabilityZonesCall.Returns.Error
}

func (c *AWSEC2Client) DescribeInstances(input *awsec2.DescribeInstancesInput) (*awsec2.DescribeInstancesOutput, error) {
	c.DescribeInstancesCall.Receives.Input = input

	return c.DescribeInstancesCall.Returns.Output, c.DescribeInstancesCall.Returns.Error
}

func (c *AWSEC2Client) DescribeVpcs(input *awsec2.DescribeVpcsInput) (*awsec2.DescribeVpcsOutput, error) {
	c.DescribeVpcsCall.Receives.Input = input

	return c.DescribeVpcsCall.Returns.Output, c.DescribeVpcsCall.Returns.Error
}
