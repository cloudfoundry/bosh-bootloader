package elb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

type Describer struct {
	elbClient Client
}

func NewDescriber(elbClient Client) Describer {
	return Describer{
		elbClient: elbClient,
	}
}

func (d Describer) Describe(elbName string) ([]string, error) {
	lbOutput, err := d.elbClient.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(elbName)},
	})
	if err != nil {
		return []string{}, err
	}

	instanceNames := []string{}

	for _, desc := range lbOutput.LoadBalancerDescriptions {
		for _, instance := range desc.Instances {
			instanceNames = append(instanceNames, aws.StringValue(instance.InstanceId))
		}
	}

	return instanceNames, nil
}
