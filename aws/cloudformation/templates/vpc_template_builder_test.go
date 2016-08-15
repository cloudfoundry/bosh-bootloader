package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VPCTemplateBuilder", func() {
	var builder templates.VPCTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewVPCTemplateBuilder()
	})

	Describe("VPC", func() {
		It("returns a template with the VPC-related parameters", func() {
			vpc := builder.VPC()

			Expect(vpc.Parameters).To(HaveLen(1))
			Expect(vpc.Parameters).To(HaveKeyWithValue("VPCCIDR", templates.Parameter{
				Description: "CIDR block for the VPC.",
				Type:        "String",
				Default:     "10.0.0.0/16",
			}))
		})

		It("returns a template with the VPC-related resources", func() {
			vpc := builder.VPC()

			Expect(vpc.Resources).To(HaveLen(3))
			Expect(vpc.Resources).To(HaveKeyWithValue("VPC", templates.Resource{
				Type: "AWS::EC2::VPC",
				Properties: templates.VPC{
					CidrBlock: templates.Ref{"VPCCIDR"},
					Tags: []templates.Tag{
						{
							Value: "bbl",
							Key:   "Name",
						},
					},
				},
			}))

			Expect(vpc.Resources).To(HaveKeyWithValue("VPCGatewayInternetGateway", templates.Resource{
				Type: "AWS::EC2::InternetGateway",
			}))

			Expect(vpc.Resources).To(HaveKeyWithValue("VPCGatewayAttachment", templates.Resource{
				Type: "AWS::EC2::VPCGatewayAttachment",
				Properties: templates.VPCGatewayAttachment{
					VpcId:             templates.Ref{"VPC"},
					InternetGatewayId: templates.Ref{"VPCGatewayInternetGateway"},
				},
			}))
		})

		It("returns a template with the VPC-related outputs", func() {
			vpc := builder.VPC()

			Expect(vpc.Outputs).To(HaveKeyWithValue("VPCID", templates.Output{
				Value: templates.Ref{Ref: "VPC"},
			}))
		})
	})
})
