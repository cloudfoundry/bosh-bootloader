package templates_test

import (
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadBalancerSubnetTemplateBuilder", func() {
	var builder templates.LoadBalancerSubnetTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewLoadBalancerSubnetTemplateBuilder()
	})

	Describe("LoadBalancerSubnet", func() {
		It("returns a template with all fields for the load balancer subnet", func() {
			subnet := builder.LoadBalancerSubnet("some-zone-1", "1", "10.0.2.0/24")

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("LoadBalancerSubnet1CIDR", templates.Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     "10.0.2.0/24",
			}))

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerSubnet1", templates.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: templates.Subnet{
					AvailabilityZone: "some-zone-1",
					CidrBlock:        templates.Ref{"LoadBalancerSubnet1CIDR"},
					VpcId:            templates.Ref{"VPC"},
					Tags: []templates.Tag{
						{
							Key:   "Name",
							Value: "LoadBalancer1",
						},
					},
				},
				DeletionPolicy: "Retain",
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerRouteTable", templates.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: templates.RouteTable{
					VpcId: templates.Ref{"VPC"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerRoute", templates.Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayAttachment",
				Properties: templates.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            templates.Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         templates.Ref{"LoadBalancerRouteTable"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerSubnet1RouteTableAssociation", templates.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: templates.SubnetRouteTableAssociation{
					RouteTableId: templates.Ref{"LoadBalancerRouteTable"},
					SubnetId:     templates.Ref{"LoadBalancerSubnet1"},
				},
			}))

			Expect(subnet.Outputs).To(HaveKeyWithValue("LoadBalancerSubnet1Name", templates.Output{
				Value: templates.Ref{
					Ref: "LoadBalancerSubnet1",
				},
			}))
		})

		It("returns subnet with az 1", func() {
			subnet := builder.LoadBalancerSubnet("some-zone-1", "1", "10.0.3.0/24")

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("LoadBalancerSubnet1CIDR", templates.Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     "10.0.3.0/24",
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerSubnet1", templates.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: templates.Subnet{
					AvailabilityZone: "some-zone-1",
					CidrBlock:        templates.Ref{"LoadBalancerSubnet1CIDR"},
					VpcId:            templates.Ref{"VPC"},
					Tags: []templates.Tag{
						{
							Key:   "Name",
							Value: "LoadBalancer1",
						},
					},
				},
				DeletionPolicy: "Retain",
			}))
		})
	})
})
