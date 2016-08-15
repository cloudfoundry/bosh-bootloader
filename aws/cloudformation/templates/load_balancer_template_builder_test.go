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
			cfRouterLoadBalancerTemplate := builder.CFRouterLoadBalancer(2, "some-certificate-arn")

			Expect(cfRouterLoadBalancerTemplate.Outputs).To(HaveLen(2))
			Expect(cfRouterLoadBalancerTemplate.Outputs).To(HaveKeyWithValue("CFRouterLoadBalancer", templates.Output{
				Value: templates.Ref{"CFRouterLoadBalancer"},
			}))

			Expect(cfRouterLoadBalancerTemplate.Outputs).To(HaveKeyWithValue("CFRouterLoadBalancerURL", templates.Output{
				Value: templates.FnGetAtt{
					[]string{
						"CFRouterLoadBalancer",
						"DNSName",
					},
				},
			}))

			Expect(cfRouterLoadBalancerTemplate.Resources).To(HaveLen(1))
			Expect(cfRouterLoadBalancerTemplate.Resources).To(HaveKeyWithValue("CFRouterLoadBalancer", templates.Resource{
				Type:      "AWS::ElasticLoadBalancing::LoadBalancer",
				DependsOn: "VPCGatewayAttachment",
				Properties: templates.ElasticLoadBalancingLoadBalancer{
					CrossZone:      true,
					Subnets:        []interface{}{templates.Ref{"LoadBalancerSubnet1"}, templates.Ref{"LoadBalancerSubnet2"}},
					SecurityGroups: []interface{}{templates.Ref{"CFRouterSecurityGroup"}},

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
						{
							Protocol:         "https",
							LoadBalancerPort: "443",
							InstanceProtocol: "http",
							InstancePort:     "80",
							SSLCertificateID: "some-certificate-arn",
						},
						{
							Protocol:         "ssl",
							LoadBalancerPort: "4443",
							InstanceProtocol: "tcp",
							InstancePort:     "80",
							SSLCertificateID: "some-certificate-arn",
						},
					},
				},
			}))
		})
	})

	Describe("CFSSHProxyLoadBalancer", func() {
		It("returns a template containing the cf ssh proxy load balancer", func() {
			cfSSHProxyLoadBalancerTemplate := builder.CFSSHProxyLoadBalancer(2)

			Expect(cfSSHProxyLoadBalancerTemplate.Outputs).To(HaveLen(2))
			Expect(cfSSHProxyLoadBalancerTemplate.Outputs).To(HaveKeyWithValue("CFSSHProxyLoadBalancer", templates.Output{
				Value: templates.Ref{"CFSSHProxyLoadBalancer"},
			}))

			Expect(cfSSHProxyLoadBalancerTemplate.Outputs).To(HaveKeyWithValue("CFSSHProxyLoadBalancerURL", templates.Output{
				Value: templates.FnGetAtt{
					[]string{
						"CFSSHProxyLoadBalancer",
						"DNSName",
					},
				},
			}))

			Expect(cfSSHProxyLoadBalancerTemplate.Resources).To(HaveLen(1))
			Expect(cfSSHProxyLoadBalancerTemplate.Resources).To(HaveKeyWithValue("CFSSHProxyLoadBalancer", templates.Resource{
				Type:      "AWS::ElasticLoadBalancing::LoadBalancer",
				DependsOn: "VPCGatewayAttachment",
				Properties: templates.ElasticLoadBalancingLoadBalancer{
					CrossZone:      true,
					Subnets:        []interface{}{templates.Ref{"LoadBalancerSubnet1"}, templates.Ref{"LoadBalancerSubnet2"}},
					SecurityGroups: []interface{}{templates.Ref{"CFSSHProxySecurityGroup"}},

					HealthCheck: templates.HealthCheck{
						HealthyThreshold:   "5",
						Interval:           "6",
						Target:             "tcp:2222",
						Timeout:            "2",
						UnhealthyThreshold: "2",
					},

					Listeners: []templates.Listener{
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

	Describe("ConcourseLoadBalancer", func() {
		It("returns a template containing the concourse load balancer", func() {
			concourseLoadBalancer := builder.ConcourseLoadBalancer(2, "some-certificate-arn")

			Expect(concourseLoadBalancer.Outputs).To(HaveLen(2))
			Expect(concourseLoadBalancer.Outputs).To(HaveKeyWithValue("ConcourseLoadBalancer", templates.Output{
				Value: templates.Ref{"ConcourseLoadBalancer"},
			}))

			Expect(concourseLoadBalancer.Outputs).To(HaveKeyWithValue("ConcourseLoadBalancerURL", templates.Output{
				Value: templates.FnGetAtt{
					[]string{
						"ConcourseLoadBalancer",
						"DNSName",
					},
				},
			}))

			Expect(concourseLoadBalancer.Resources).To(HaveLen(1))
			Expect(concourseLoadBalancer.Resources).To(HaveKeyWithValue("ConcourseLoadBalancer", templates.Resource{
				Type:      "AWS::ElasticLoadBalancing::LoadBalancer",
				DependsOn: "VPCGatewayAttachment",
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
						{
							Protocol:         "ssl",
							LoadBalancerPort: "443",
							InstanceProtocol: "tcp",
							InstancePort:     "8080",
							SSLCertificateID: "some-certificate-arn",
						},
					},
				},
			}))
		})
	})
})
