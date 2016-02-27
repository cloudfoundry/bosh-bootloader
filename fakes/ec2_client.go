package fakes

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type EC2Client struct {
	ImportKeyPairCall struct {
		Receives struct {
			Input *ec2.ImportKeyPairInput
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
			DescribeKeyPairsOutput *awsec2.DescribeKeyPairsOutput
			Error                  error
		}
	}
}

func (c *EC2Client) ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error) {
	c.ImportKeyPairCall.Receives.Input = input

	return nil, c.ImportKeyPairCall.Returns.Error
}

func (c *EC2Client) DescribeKeyPairs(input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
	c.DescribeKeyPairsCall.Receives.Input = input

	return c.DescribeKeyPairsCall.Returns.DescribeKeyPairsOutput, c.DescribeKeyPairsCall.Returns.Error
}
