package templates

import "fmt"

type WebELBTemplateBuilder struct{}

func NewWebELBTemplateBuilder() WebELBTemplateBuilder {
	return WebELBTemplateBuilder{}
}

func (t WebELBTemplateBuilder) WebELBLoadBalancer(numberOfAvailabliltyZones int) Template {
	subnets := []interface{}{}
	for i := 1; i <= numberOfAvailabliltyZones; i++ {
		subnets = append(subnets, Ref{fmt.Sprintf("LoadBalancerSubnet%d", i)})
	}

	return Template{
		Outputs: map[string]Output{
			"LB": {Value: Ref{"WebELBLoadBalancer"}},
			"LBURL": {
				Value: FnGetAtt{
					[]string{
						"WebELBLoadBalancer",
						"DNSName",
					},
				},
			},
		},
		Resources: map[string]Resource{
			"WebELBLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					Subnets:        subnets,
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
