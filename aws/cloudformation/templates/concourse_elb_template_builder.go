package templates

import "fmt"

type ConcourseELBTemplateBuilder struct{}

func NewConcourseELBTemplateBuilder() ConcourseELBTemplateBuilder {
	return ConcourseELBTemplateBuilder{}
}

func (t ConcourseELBTemplateBuilder) ConcourseLoadBalancer(numberOfAvailabliltyZones int) Template {
	subnets := []interface{}{}
	for i := 1; i <= numberOfAvailabliltyZones; i++ {
		subnets = append(subnets, Ref{fmt.Sprintf("LoadBalancerSubnet%d", i)})
	}

	return Template{
		Outputs: map[string]Output{
			"LB": {Value: Ref{"ConcourseLoadBalancer"}},
			"LBURL": {
				Value: FnGetAtt{
					[]string{
						"ConcourseLoadBalancer",
						"DNSName",
					},
				},
			},
		},
		Resources: map[string]Resource{
			"ConcourseLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					Subnets:        subnets,
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
