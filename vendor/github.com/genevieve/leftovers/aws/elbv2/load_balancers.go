package elbv2

import (
	"fmt"
	"strings"

	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevieve/leftovers/common"
)

type loadBalancersClient interface {
	DescribeLoadBalancers(*awselbv2.DescribeLoadBalancersInput) (*awselbv2.DescribeLoadBalancersOutput, error)
	DeleteLoadBalancer(*awselbv2.DeleteLoadBalancerInput) (*awselbv2.DeleteLoadBalancerOutput, error)
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
	loadBalancers, err := l.client.DescribeLoadBalancers(&awselbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("Describe ELBV2 Load Balancers: %s", err)

	}

	var resources []common.Deletable
	for _, lb := range loadBalancers.LoadBalancers {
		r := NewLoadBalancer(l.client, lb.LoadBalancerName, lb.LoadBalancerArn)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := l.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (l LoadBalancers) Type() string {
	return "elbv2-load-balancer"
}
