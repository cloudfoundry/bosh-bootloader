package templates

import "fmt"

type LoadBalancerSubnetsTemplateBuilder struct{}

func NewLoadBalancerSubnetsTemplateBuilder() LoadBalancerSubnetsTemplateBuilder {
	return LoadBalancerSubnetsTemplateBuilder{}
}

func (LoadBalancerSubnetsTemplateBuilder) LoadBalancerSubnets(azCount int, envID string) Template {
	loadBalancerSubnetTemplateBuilder := NewLoadBalancerSubnetTemplateBuilder()

	template := Template{}
	for index := 1; index <= azCount; index++ {
		template = template.Merge(loadBalancerSubnetTemplateBuilder.LoadBalancerSubnet(
			index-1,
			fmt.Sprintf("%d", index),
			fmt.Sprintf("10.0.%d.0/24", index+1),
			envID,
		))
	}
	return template
}
