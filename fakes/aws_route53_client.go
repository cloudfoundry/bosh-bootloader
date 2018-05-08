package fakes

import (
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
)

type AWSRoute53Client struct {
	ListHostedZonesByNameCall struct {
		Receives struct {
			Input *awsroute53.ListHostedZonesByNameInput
		}
		Returns struct {
			Output *awsroute53.ListHostedZonesByNameOutput
			Error  error
		}
	}

	GetHostedZoneCall struct {
		CallCount int
		Receives  struct {
			Input *awsroute53.GetHostedZoneInput
		}
		Returns struct {
			Output *awsroute53.GetHostedZoneOutput
			Error  error
		}
	}
}

func (c *AWSRoute53Client) ListHostedZonesByName(input *awsroute53.ListHostedZonesByNameInput) (*awsroute53.ListHostedZonesByNameOutput, error) {
	c.ListHostedZonesByNameCall.Receives.Input = input

	return c.ListHostedZonesByNameCall.Returns.Output, c.ListHostedZonesByNameCall.Returns.Error
}

func (c *AWSRoute53Client) GetHostedZone(input *awsroute53.GetHostedZoneInput) (*awsroute53.GetHostedZoneOutput, error) {
	c.GetHostedZoneCall.CallCount++
	c.GetHostedZoneCall.Receives.Input = input

	return c.GetHostedZoneCall.Returns.Output, c.GetHostedZoneCall.Returns.Error
}
