package templates

import "fmt"

type LoadBalancerSubnetsTemplateBuilder struct{}

func NewLoadBalancerSubnetsTemplateBuilder() LoadBalancerSubnetsTemplateBuilder {
	return LoadBalancerSubnetsTemplateBuilder{}
}

func (LoadBalancerSubnetsTemplateBuilder) LoadBalancerSubnets(availabilityZones []string) Template {
	loadBalancerSubnetTemplateBuilder := NewLoadBalancerSubnetTemplateBuilder()

	template := Template{}
	for index, az := range availabilityZones {
		template = template.Merge(loadBalancerSubnetTemplateBuilder.LoadBalancerSubnet(
			az,
			fmt.Sprintf("%d", index+1),
			fmt.Sprintf("10.0.%d.0/24", index+2),
		))
	}
	return template
}
