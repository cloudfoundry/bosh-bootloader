package templates

import "fmt"

type CFLoadBalancerTemplateBuilder struct{}

func NewCFLoadBalancerTemplateBuilder() CFLoadBalancerTemplateBuilder {
	return CFLoadBalancerTemplateBuilder{}
}

func (CFLoadBalancerTemplateBuilder) CFLoadBalancer(numberOfAvailabliltyZones int) Template {
	subnets := []interface{}{}
	for i := 1; i <= numberOfAvailabliltyZones; i++ {
		subnets = append(subnets, Ref{fmt.Sprintf("LoadBalancerSubnet%d", i)})
	}

	return Template{
		Outputs: map[string]Output{
			"CFLB": {Value: Ref{"CFLoadBalancer"}},
			"CFLBURL": {
				Value: FnGetAtt{
					[]string{
						"CFLoadBalancer",
						"DNSName",
					},
				},
			},
		},
		Resources: map[string]Resource{
			"CFLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					CrossZone:      true,
					Subnets:        subnets,
					SecurityGroups: []interface{}{Ref{"CFSecurityGroup"}},

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
