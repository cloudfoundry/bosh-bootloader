package fakes

import "github.com/aws/aws-sdk-go/service/elb"

type ELBClient struct {
	DescribeLoadBalancersCall struct {
		Receives struct {
			Input *elb.DescribeLoadBalancersInput
		}
		Returns struct {
			Output *elb.DescribeLoadBalancersOutput
			Error  error
		}
	}
}

func (e *ELBClient) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	e.DescribeLoadBalancersCall.Receives.Input = input

	return e.DescribeLoadBalancersCall.Returns.Output, e.DescribeLoadBalancersCall.Returns.Error
}
