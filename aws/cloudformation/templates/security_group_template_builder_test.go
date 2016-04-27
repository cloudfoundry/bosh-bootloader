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
					SecurityGroupEgress: []string{},
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
					SecurityGroupEgress: []string{},
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

	Describe("ConcourseSecurityGroup", func() {
		It("returns a template", func() {
			securityGroup := builder.ConcourseSecurityGroup()

			Expect(securityGroup.Resources).To(HaveLen(3))
			Expect(securityGroup.Resources).To(HaveKeyWithValue("ConcourseSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "Concourse",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []templates.SecurityGroupIngress{
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "80",
							ToPort:     "80",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "2222",
							ToPort:     "2222",
						},
					},
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromConcourseSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"ConcourseSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromConcourseSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"ConcourseSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))
		})
	})

	Describe("CFRouterSecurityGroup", func() {
		It("returns a template containing the router security group", func() {
			securityGroup := builder.CFRouterSecurityGroup()

			Expect(securityGroup.Resources).To(HaveLen(3))
			Expect(securityGroup.Resources).To(HaveKeyWithValue("CFRouterSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "Router",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []templates.SecurityGroupIngress{
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "80",
							ToPort:     "80",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "2222",
							ToPort:     "2222",
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

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromCFRouterSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"CFRouterSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromCFRouterSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"CFRouterSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))
		})
	})

	Describe("CFSSHProxySecurityGroup", func() {
		It("returns a template containing the cf ssh proxy security group", func() {
			securityGroup := builder.CFSSHProxySecurityGroup()

			Expect(securityGroup.Resources).To(HaveLen(2))
			Expect(securityGroup.Resources).To(HaveKeyWithValue("CFSSHProxySecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "CFSSHProxy",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []templates.SecurityGroupIngress{
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "2222",
							ToPort:     "2222",
						},
					},
				},
			}))

			Expect(securityGroup.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromCFSSHProxySecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"CFSSHProxySecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))
		})
	})
})
