package elb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

type Describer struct{}

func NewDescriber() Describer {
	return Describer{}
}

func (d Describer) Describe(elbName string, client Client) ([]string, error) {
	lbOutput, err := client.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
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
