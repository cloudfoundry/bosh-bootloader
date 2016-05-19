package elb

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"

	awselb "github.com/aws/aws-sdk-go/service/elb"
)

type Client interface {
	DescribeLoadBalancers(input *awselb.DescribeLoadBalancersInput) (*awselb.DescribeLoadBalancersOutput, error)
}

func NewClient(config aws.Config) Client {
	return awselb.New(session.New(config.ClientConfig()))
}
