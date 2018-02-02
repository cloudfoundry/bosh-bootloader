package elb

import (
	"fmt"
	"strings"

	awselb "github.com/aws/aws-sdk-go/service/elb"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type loadBalancersClient interface {
	DescribeLoadBalancers(*awselb.DescribeLoadBalancersInput) (*awselb.DescribeLoadBalancersOutput, error)
	DeleteLoadBalancer(*awselb.DeleteLoadBalancerInput) (*awselb.DeleteLoadBalancerOutput, error)
}

type LoadBalancers struct {
	client loadBalancersClient
	logger logger
}

func NewLoadBalancers(client loadBalancersClient, logger logger) LoadBalancers {
	return LoadBalancers{
		client: client,
		logger: logger,
	}
}

func (l LoadBalancers) List(filter string) ([]common.Deletable, error) {
	loadBalancers, err := l.client.DescribeLoadBalancers(&awselb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing load balancers: %s", err)
	}

	var resources []common.Deletable
	for _, lb := range loadBalancers.LoadBalancerDescriptions {
		resource := NewLoadBalancer(l.client, lb.LoadBalancerName)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := l.logger.Prompt(fmt.Sprintf("Are you sure you want to delete load balancer %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
