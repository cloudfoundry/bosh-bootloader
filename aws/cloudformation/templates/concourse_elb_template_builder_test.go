package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConcourseELBTemplateBuilder", func() {
	var builder templates.ConcourseELBTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewConcourseELBTemplateBuilder()
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
