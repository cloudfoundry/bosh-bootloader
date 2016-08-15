package templates_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

var _ = Describe("SecurityGroupTemplateBuilder", func() {
	var builder templates.SecurityGroupTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewSecurityGroupTemplateBuilder()
	})

	Describe("InternalSecurityGroup", func() {
		It("returns a template containing all the fields for internal security group", func() {
			securityGroup := builder.InternalSecurityGroup()

			Expect(securityGroup.Resources).To(HaveLen(5))
			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "Internal",
					SecurityGroupEgress: []templates.SecurityGroupEgress{},
					SecurityGroupIngress: []templates.SecurityGroupIngress{
						{
							IpProtocol: "tcp",
							FromPort:   "0",
							ToPort:     "65535",
						},
						{
							IpProtocol: "udp",
							FromPort:   "0",
							ToPort:     "65535",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "icmp",
							FromPort:   "-1",
							ToPort:     "-1",
						},
					},
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromBOSH", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"BOSHSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromBOSH", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"BOSHSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromSelf", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"InternalSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromSelf", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"InternalSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(securityGroup.Outputs).To(HaveLen(1))
			Expect(securityGroup.Outputs).To(HaveKeyWithValue("InternalSecurityGroup", templates.Output{
				Value: templates.Ref{"InternalSecurityGroup"},
			}))
		})
	})

	Describe("BOSHSecurityGroup", func() {
		It("returns a template containing the bosh security group", func() {
			securityGroup := builder.BOSHSecurityGroup()

			Expect(securityGroup.Parameters).To(HaveLen(1))
			Expect(securityGroup.Parameters).To(HaveKeyWithValue("BOSHInboundCIDR", templates.Parameter{
				Description: "CIDR to permit access to BOSH (e.g. 205.103.216.37/32 for your specific IP)",
				Type:        "String",
				Default:     "0.0.0.0/0",
			}))

			Expect(securityGroup.Outputs).To(HaveLen(1))
			Expect(securityGroup.Outputs).To(HaveKeyWithValue("BOSHSecurityGroup", templates.Output{
				Value: templates.Ref{"BOSHSecurityGroup"},
			}))

			Expect(securityGroup.Resources).To(HaveLen(1))
			Expect(securityGroup.Resources).To(HaveKeyWithValue("BOSHSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "BOSH",
					SecurityGroupEgress: []templates.SecurityGroupEgress{},
					SecurityGroupIngress: []templates.SecurityGroupIngress{
						{
							CidrIp:     templates.Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "22",
							ToPort:     "22",
						},

						{
							CidrIp:     templates.Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "6868",
							ToPort:     "6868",
						},
						{
							CidrIp:     templates.Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "25555",
							ToPort:     "25555",
						},
						{
							SourceSecurityGroupId: templates.Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
						{
							SourceSecurityGroupId: templates.Ref{"InternalSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
				},
			}))
		})
	})

	Context("when building security groups for load balancers", func() {
		var (
			loadBalancerTemplate templates.Template
		)

		BeforeEach(func() {
			loadBalancerTemplate = templates.Template{
				Resources: map[string]templates.Resource{
					"some-load-balancer": {
						DependsOn: "VPCGatewayAttachment",
						Type:      "AWS::ElasticLoadBalancing::LoadBalancer",
						Properties: templates.ElasticLoadBalancingLoadBalancer{
							Listeners: []templates.Listener{
								{
									Protocol:         "tcp",
									LoadBalancerPort: "1000",
									InstanceProtocol: "tcp",
									InstancePort:     "2222",
								},
								{
									Protocol:         "http",
									LoadBalancerPort: "80",
									InstanceProtocol: "http",
									InstancePort:     "8080",
								},
								{
									Protocol:         "https",
									LoadBalancerPort: "4443",
									InstanceProtocol: "tcp",
									InstancePort:     "8080",
								},
								{
									Protocol:         "ssl",
									LoadBalancerPort: "443",
									InstanceProtocol: "tcp",
									InstancePort:     "8080",
								},
							},
						},
					},
				},
			}
		})

		Describe("LBSecurityGroup", func() {
			It("returns a load balancer security group based on load balancer template", func() {
				securityGroup := builder.LBSecurityGroup("some-security-group", "some-group-description",
					"some-load-balancer", loadBalancerTemplate)

				Expect(securityGroup.Resources).To(HaveLen(1))
				Expect(securityGroup.Resources).To(HaveKeyWithValue("some-security-group", templates.Resource{
					Type: "AWS::EC2::SecurityGroup",
					Properties: templates.SecurityGroup{
						VpcId:               templates.Ref{"VPC"},
						GroupDescription:    "some-group-description",
						SecurityGroupEgress: []templates.SecurityGroupEgress{},
						SecurityGroupIngress: []templates.SecurityGroupIngress{
							{
								CidrIp:     "0.0.0.0/0",
								IpProtocol: "tcp",
								FromPort:   "1000",
								ToPort:     "1000",
							},
							{
								CidrIp:     "0.0.0.0/0",
								IpProtocol: "tcp",
								FromPort:   "80",
								ToPort:     "80",
							},
							{
								CidrIp:     "0.0.0.0/0",
								IpProtocol: "tcp",
								FromPort:   "4443",
								ToPort:     "4443",
							},
							{
								CidrIp:     "0.0.0.0/0",
								IpProtocol: "tcp",
								FromPort:   "443",
								ToPort:     "443",
							},
						},
					},
				}))
			})
		})

		Describe("LBInternalSecurityGroup", func() {
			It("returns a load balancer internal security group based on load balancer template", func() {
				securityGroup := builder.LBInternalSecurityGroup("some-internal-security-group", "some-security-group",
					"some-group-description", "some-load-balancer", loadBalancerTemplate)

				Expect(securityGroup.Resources).To(HaveLen(1))
				Expect(securityGroup.Resources).To(HaveKeyWithValue("some-internal-security-group", templates.Resource{
					Type: "AWS::EC2::SecurityGroup",
					Properties: templates.SecurityGroup{
						VpcId:               templates.Ref{"VPC"},
						GroupDescription:    "some-group-description",
						SecurityGroupEgress: []templates.SecurityGroupEgress{},
						SecurityGroupIngress: []templates.SecurityGroupIngress{
							{
								SourceSecurityGroupId: templates.Ref{"some-security-group"},
								IpProtocol:            "tcp",
								FromPort:              "2222",
								ToPort:                "2222",
							},
							{
								SourceSecurityGroupId: templates.Ref{"some-security-group"},
								IpProtocol:            "tcp",
								FromPort:              "8080",
								ToPort:                "8080",
							},
						},
					},
				}))

				Expect(securityGroup.Outputs).To(HaveLen(1))
				Expect(securityGroup.Outputs).To(HaveKeyWithValue("some-internal-security-group", templates.Output{
					Value: templates.Ref{"some-internal-security-group"},
				}))
			})
		})
	})
})
