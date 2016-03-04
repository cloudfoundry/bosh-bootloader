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

	Describe("Internal Security Group", func() {
		It("returns a template containing all the fields for internal security group", func() {
			security_group := builder.InternalSecurityGroup()

			Expect(security_group.Resources).To(HaveLen(5))
			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "Internal",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []templates.SecurityGroupIngress{
						{
							SourceSecurityGroupId: templates.Ref{"WebSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
						{
							SourceSecurityGroupId: templates.Ref{"WebSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
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

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromBOSH", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"BOSHSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromBOSH", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"BOSHSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromSelf", templates.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: templates.SecurityGroupIngress{
					GroupId:               templates.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: templates.Ref{"InternalSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromSelf", templates.Resource{
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

	Describe("BOSH Security Group", func() {
		It("returns a template containing the bosh security group", func() {
			security_group := builder.BOSHSecurityGroup()

			Expect(security_group.Parameters).To(HaveLen(1))
			Expect(security_group.Parameters).To(HaveKeyWithValue("BOSHInboundCIDR", templates.Parameter{
				Description: "CIDR to permit access to BOSH (e.g. 205.103.216.37/32 for your specific IP)",
				Type:        "String",
				Default:     "0.0.0.0/0",
			}))

			Expect(security_group.Resources).To(HaveLen(1))
			Expect(security_group.Resources).To(HaveKeyWithValue("BOSHSecurityGroup", templates.Resource{
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

	Describe("Web Security Group", func() {
		It("returns a template containing the web security group", func() {
			security_group := builder.WebSecurityGroup()

			Expect(security_group.Resources).To(HaveLen(1))
			Expect(security_group.Resources).To(HaveKeyWithValue("WebSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "Web",
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
		})
	})
})
