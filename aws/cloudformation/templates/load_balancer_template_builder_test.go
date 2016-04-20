package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadBalancerTemplateBuilder", func() {
	var builder templates.LoadBalancerTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewLoadBalancerTemplateBuilder()
	})

	Describe("CFRouterLoadBalancer", func() {
		It("returns a template containing the cf load balancer", func() {
			cfRouterLoadBalancerTemplate := builder.CFRouterLoadBalancer(2)

			Expect(cfRouterLoadBalancerTemplate.Outputs).To(HaveLen(2))
			Expect(cfRouterLoadBalancerTemplate.Outputs).To(HaveKeyWithValue("CFLB", templates.Output{
				Value: templates.Ref{"CFRouterLoadBalancer"},
			}))

			Expect(cfRouterLoadBalancerTemplate.Outputs).To(HaveKeyWithValue("CFLBURL", templates.Output{
				Value: templates.FnGetAtt{
					[]string{
						"CFRouterLoadBalancer",
						"DNSName",
					},
				},
			}))

			Expect(cfRouterLoadBalancerTemplate.Resources).To(HaveLen(1))
			Expect(cfRouterLoadBalancerTemplate.Resources).To(HaveKeyWithValue("CFRouterLoadBalancer", templates.Resource{
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: templates.ElasticLoadBalancingLoadBalancer{
					CrossZone:      true,
					Subnets:        []interface{}{templates.Ref{"LoadBalancerSubnet1"}, templates.Ref{"LoadBalancerSubnet2"}},
					SecurityGroups: []interface{}{templates.Ref{"CFSecurityGroup"}},

					HealthCheck: templates.HealthCheck{
						HealthyThreshold:   "5",
						Interval:           "12",
						Target:             "tcp:80",
						Timeout:            "2",
						UnhealthyThreshold: "2",
					},

					Listeners: []templates.Listener{
						{
							Protocol:         "http",
							LoadBalancerPort: "80",
							InstanceProtocol: "http",
							InstancePort:     "80",
						},
					},
				},
			}))
		})
	})

	Describe("ConcourseLoadBalancer", func() {
		It("returns a template containing the web elb load balancer", func() {
			concourseLoadBalancer := builder.ConcourseLoadBalancer(2)

			Expect(concourseLoadBalancer.Outputs).To(HaveLen(2))
			Expect(concourseLoadBalancer.Outputs).To(HaveKeyWithValue("LB", templates.Output{
				Value: templates.Ref{"ConcourseLoadBalancer"},
			}))

			Expect(concourseLoadBalancer.Outputs).To(HaveKeyWithValue("LBURL", templates.Output{
				Value: templates.FnGetAtt{
					[]string{
						"ConcourseLoadBalancer",
						"DNSName",
					},
				},
			}))

			Expect(concourseLoadBalancer.Resources).To(HaveLen(1))
			Expect(concourseLoadBalancer.Resources).To(HaveKeyWithValue("ConcourseLoadBalancer", templates.Resource{
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: templates.ElasticLoadBalancingLoadBalancer{
					Subnets:        []interface{}{templates.Ref{"LoadBalancerSubnet1"}, templates.Ref{"LoadBalancerSubnet2"}},
					SecurityGroups: []interface{}{templates.Ref{"ConcourseSecurityGroup"}},

					HealthCheck: templates.HealthCheck{
						HealthyThreshold:   "2",
						Interval:           "30",
						Target:             "tcp:8080",
						Timeout:            "5",
						UnhealthyThreshold: "10",
					},

					Listeners: []templates.Listener{
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
			}))
		})
	})
})
