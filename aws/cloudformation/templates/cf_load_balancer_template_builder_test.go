package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CFLoadBalancerTemplateBuilder", func() {
	var builder templates.CFLoadBalancerTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewCFLoadBalancerTemplateBuilder()
	})

	Describe("CFLoadBalancer", func() {
		It("returns a template containing the cf load balancer", func() {
			cf_load_balancer := builder.CFLoadBalancer(2)

			Expect(cf_load_balancer.Outputs).To(HaveLen(2))
			Expect(cf_load_balancer.Outputs).To(HaveKeyWithValue("CFLB", templates.Output{
				Value: templates.Ref{"CFLoadBalancer"},
			}))

			Expect(cf_load_balancer.Outputs).To(HaveKeyWithValue("CFLBURL", templates.Output{
				Value: templates.FnGetAtt{
					[]string{
						"CFLoadBalancer",
						"DNSName",
					},
				},
			}))

			Expect(cf_load_balancer.Resources).To(HaveLen(1))
			Expect(cf_load_balancer.Resources).To(HaveKeyWithValue("CFLoadBalancer", templates.Resource{
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
})
