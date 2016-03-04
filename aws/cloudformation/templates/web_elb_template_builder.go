package templates

type WebELBTemplateBuilder struct{}

func NewWebELBTemplateBuilder() WebELBTemplateBuilder {
	return WebELBTemplateBuilder{}
}

func (t WebELBTemplateBuilder) WebELBLoadBalancer() Template {
	return Template{
		Resources: map[string]Resource{
			"WebELBLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					Subnets:        []interface{}{Ref{"LoadBalancerSubnet"}},
					SecurityGroups: []interface{}{Ref{"WebSecurityGroup"}},

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
