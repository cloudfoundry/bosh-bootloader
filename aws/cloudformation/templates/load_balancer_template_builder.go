package templates

import "fmt"

type LoadBalancerTemplateBuilder struct{}

func NewLoadBalancerTemplateBuilder() LoadBalancerTemplateBuilder {
	return LoadBalancerTemplateBuilder{}
}

func (l LoadBalancerTemplateBuilder) CFSSHProxyLoadBalancer(numberOfAvailabilityZones int) Template {
	return Template{
		Outputs: l.outputsFor("CFSSHProxyLoadBalancer"),
		Resources: map[string]Resource{
			"CFSSHProxyLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					CrossZone:      true,
					Subnets:        l.loadBalancerSubnets(numberOfAvailabilityZones),
					SecurityGroups: []interface{}{Ref{"CFSSHProxySecurityGroup"}},

					HealthCheck: HealthCheck{
						HealthyThreshold:   "5",
						Interval:           "6",
						Target:             "tcp:2222",
						Timeout:            "2",
						UnhealthyThreshold: "2",
					},

					Listeners: []Listener{
						{
							Protocol:         "tcp",
							LoadBalancerPort: "2222",
							InstanceProtocol: "tcp",
							InstancePort:     "2222",
						},
					},
				},
			},
		},
	}
}

func (l LoadBalancerTemplateBuilder) CFRouterLoadBalancer(numberOfAvailabilityZones int) Template {
	return Template{
		Outputs: l.outputsFor("CFRouterLoadBalancer"),
		Resources: map[string]Resource{
			"CFRouterLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					CrossZone:      true,
					Subnets:        l.loadBalancerSubnets(numberOfAvailabilityZones),
					SecurityGroups: []interface{}{Ref{"CFRouterSecurityGroup"}},

					HealthCheck: HealthCheck{
						HealthyThreshold:   "5",
						Interval:           "12",
						Target:             "tcp:80",
						Timeout:            "2",
						UnhealthyThreshold: "2",
					},

					Listeners: []Listener{
						{
							Protocol:         "http",
							LoadBalancerPort: "80",
							InstanceProtocol: "http",
							InstancePort:     "80",
						},
					},
				},
			},
		},
	}
}

func (l LoadBalancerTemplateBuilder) ConcourseLoadBalancer(numberOfAvailabilityZones int) Template {
	return Template{
		Outputs: l.outputsFor("ConcourseLoadBalancer"),
		Resources: map[string]Resource{
			"ConcourseLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					Subnets:        l.loadBalancerSubnets(numberOfAvailabilityZones),
					SecurityGroups: []interface{}{Ref{"ConcourseSecurityGroup"}},

					HealthCheck: HealthCheck{
						HealthyThreshold:   "2",
						Interval:           "30",
						Target:             "tcp:8080",
						Timeout:            "5",
						UnhealthyThreshold: "10",
					},

					Listeners: []Listener{
						{
							Protocol:         "tcp",
							LoadBalancerPort: "80",
							InstanceProtocol: "tcp",
							InstancePort:     "8080",
						},
						{
							Protocol:         "tcp",
							LoadBalancerPort: "2222",
							InstanceProtocol: "tcp",
							InstancePort:     "2222",
						},
					},
				},
			},
		},
	}
}

func (LoadBalancerTemplateBuilder) outputsFor(loadBalancerName string) map[string]Output {
	return map[string]Output{
		loadBalancerName: {Value: Ref{loadBalancerName}},
		loadBalancerName + "URL": {
			Value: FnGetAtt{
				[]string{
					loadBalancerName,
					"DNSName",
				},
			},
		},
	}
}

func (LoadBalancerTemplateBuilder) loadBalancerSubnets(numberOfAvailabilityZones int) []interface{} {
	subnets := []interface{}{}
	for i := 1; i <= numberOfAvailabilityZones; i++ {
		subnets = append(subnets, Ref{fmt.Sprintf("LoadBalancerSubnet%d", i)})
	}

	return subnets
}
