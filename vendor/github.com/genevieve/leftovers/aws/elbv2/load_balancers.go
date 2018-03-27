package elbv2

import (
	"fmt"
	"strings"

	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevieve/leftovers/aws/common"
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

func (l LoadBalancers) ListAll(filter string) ([]common.Deletable, error) {
	return l.get(filter)
}

func (l LoadBalancers) List(filter string) ([]common.Deletable, error) {
	resources, err := l.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := l.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (l LoadBalancers) get(filter string) ([]common.Deletable, error) {
	loadBalancers, err := l.client.DescribeLoadBalancers(&awselbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing load balancers: %s", err)
	}

	var resources []common.Deletable
	for _, lb := range loadBalancers.LoadBalancers {
		resource := NewLoadBalancer(l.client, lb.LoadBalancerName, lb.LoadBalancerArn)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
