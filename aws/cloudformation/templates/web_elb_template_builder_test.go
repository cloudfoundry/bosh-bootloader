package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WebELBTemplateBuilder", func() {
	var builder templates.WebELBTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewWebELBTemplateBuilder()
	})

	Describe("Web ELB Load Balancer", func() {
		It("returns a template containing the web elb load balancer", func() {
			web_elb_load_balancer := builder.WebELBLoadBalancer()

			Expect(web_elb_load_balancer.Outputs).To(HaveLen(2))
			Expect(web_elb_load_balancer.Outputs).To(HaveKeyWithValue("LB", templates.Output{
				Value: templates.Ref{"WebELBLoadBalancer"},
			}))

			Expect(web_elb_load_balancer.Outputs).To(HaveKeyWithValue("LBURL", templates.Output{
				Value: templates.FnGetAtt{
					[]string{
						"WebELBLoadBalancer",
						"DNSName",
					},
				},
			}))

			Expect(web_elb_load_balancer.Resources).To(HaveLen(1))
			Expect(web_elb_load_balancer.Resources).To(HaveKeyWithValue("WebELBLoadBalancer", templates.Resource{
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: templates.ElasticLoadBalancingLoadBalancer{
					Subnets:        []interface{}{templates.Ref{"LoadBalancerSubnet"}},
					SecurityGroups: []interface{}{templates.Ref{"WebSecurityGroup"}},

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
