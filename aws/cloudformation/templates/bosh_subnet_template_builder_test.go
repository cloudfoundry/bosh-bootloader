package templates_test

import (
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSHSubnetTemplateBuilder", func() {
	var builder templates.BOSHSubnetTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewBOSHSubnetTemplateBuilder()
	})

	Describe("BOSHSubnet", func() {
		It("returns a template with all fields for the BOSH subnet", func() {
			subnet := builder.BOSHSubnet("some-availability-zone")

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHSubnet", templates.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: templates.Subnet{
					VpcId:            templates.Ref{"VPC"},
					CidrBlock:        templates.Ref{"BOSHSubnetCIDR"},
					AvailabilityZone: "some-availability-zone",
					Tags: []templates.Tag{
						{
							Key:   "Name",
							Value: "BOSH",
						},
					},
				},
				DeletionPolicy: "Retain",
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHRouteTable", templates.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: templates.RouteTable{
					VpcId: templates.Ref{"VPC"},
				},
				DeletionPolicy: "Retain",
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHRoute", templates.Resource{
				DependsOn: "VPCGatewayAttachment",
				Type:      "AWS::EC2::Route",
				Properties: templates.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            templates.Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         templates.Ref{"BOSHRouteTable"},
				},
				DeletionPolicy: "Retain",
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHSubnetRouteTableAssociation", templates.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: templates.SubnetRouteTableAssociation{
					RouteTableId: templates.Ref{"BOSHRouteTable"},
					SubnetId:     templates.Ref{"BOSHSubnet"},
				},
				DeletionPolicy: "Retain",
			}))

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("BOSHSubnetCIDR", templates.Parameter{
				Description: "CIDR block for the BOSH subnet.",
				Type:        "String",
				Default:     "10.0.0.0/24",
			}))

			Expect(subnet.Outputs).To(HaveLen(2))
			Expect(subnet.Outputs).To(HaveKeyWithValue("BOSHSubnet", templates.Output{
				Value: templates.Ref{"BOSHSubnet"},
			}))
			Expect(subnet.Outputs).To(HaveKeyWithValue("BOSHRouteTable", templates.Output{
				Value: templates.Ref{"BOSHRouteTable"},
			}))
		})
	})
})
