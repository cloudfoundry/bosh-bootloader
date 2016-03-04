package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NATTemplateBuilder", func() {
	var builder templates.NATTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewNATTemplateBuilder()
	})

	Describe("NAT", func() {
		It("returns a template containing all of the NAT fields", func() {
			nat := builder.NAT()

			Expect(nat.Mappings).To(HaveLen(1))
			Expect(nat.Mappings).To(HaveKeyWithValue("AWSNATAMI", map[string]templates.AMI{
				"us-east-1":      {"ami-68115b02"},
				"us-west-1":      {"ami-ef1a718f"},
				"us-west-2":      {"ami-77a4b816"},
				"eu-west-1":      {"ami-c0993ab3"},
				"eu-central-1":   {"ami-0b322e67"},
				"ap-southeast-1": {"ami-e2fc3f81"},
				"ap-southeast-2": {"ami-e3217a80"},
				"ap-northeast-1": {"ami-f885ae96"},
				"ap-northeast-2": {"ami-4118d72f"},
				"sa-east-1":      {"ami-8631b5ea"},
			}))

			Expect(nat.Resources).To(HaveLen(3))
			Expect(nat.Resources).To(HaveKeyWithValue("NATSecurityGroup", templates.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: templates.SecurityGroup{
					VpcId:               templates.Ref{"VPC"},
					GroupDescription:    "NAT",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []templates.SecurityGroupIngress{
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

			Expect(nat.Resources).To(HaveKeyWithValue("NATInstance", templates.Resource{
				Type: "AWS::EC2::Instance",
				Properties: templates.Instance{
					InstanceType:    "m4.large",
					SubnetId:        templates.Ref{"BOSHSubnet"},
					SourceDestCheck: false,
					ImageId: map[string]interface{}{
						"Fn::FindInMap": []interface{}{
							"AWSNATAMI",
							templates.Ref{"AWS::Region"},
							"AMI",
						},
					},
					KeyName: templates.Ref{"SSHKeyPairName"},
					SecurityGroupIds: []interface{}{
						templates.Ref{"NATSecurityGroup"},
					},
					Tags: []templates.Tag{
						{
							Key:   "Name",
							Value: "NAT",
						},
					},
				},
			}))
			Expect(nat.Resources).To(HaveKeyWithValue("NATEIP", templates.Resource{
				Type: "AWS::EC2::EIP",
				Properties: templates.EIP{
					Domain:     "vpc",
					InstanceId: templates.Ref{"NATInstance"},
				},
			}))
		})
	})
})
