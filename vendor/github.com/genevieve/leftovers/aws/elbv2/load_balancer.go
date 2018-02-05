package elbv2

import (
	"fmt"

	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
)

type LoadBalancer struct {
	client     loadBalancersClient
	name       *string
	arn        *string
	identifier string
}

func NewLoadBalancer(client loadBalancersClient, name, arn *string) LoadBalancer {
	return LoadBalancer{
		client:     client,
		name:       name,
		arn:        arn,
		identifier: *name,
	}
}

func (l LoadBalancer) Delete() error {
	_, err := l.client.DeleteLoadBalancer(&awselbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: l.arn,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting load balancer %s: %s", l.identifier, err)
	}

	return nil
}

func (l LoadBalancer) Name() string {
	return l.identifier
}
