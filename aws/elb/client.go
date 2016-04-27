package elb

import awselb "github.com/aws/aws-sdk-go/service/elb"

type Client interface {
	DescribeLoadBalancers(input *awselb.DescribeLoadBalancersInput) (*awselb.DescribeLoadBalancersOutput, error)
}
