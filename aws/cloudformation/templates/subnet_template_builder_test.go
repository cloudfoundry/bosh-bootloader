package templates_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

var _ = Describe("SubnetTemplateBuilder", func() {
	var builder templates.SubnetTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewSubnetTemplateBuilder()
	})

	Describe("BOSHSubnet", func() {
		It("returns a template with all fields for the BOSH subnet", func() {
			subnet := builder.BOSHSubnet()

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHSubnet", templates.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: templates.Subnet{
					VpcId:     templates.Ref{"VPC"},
					CidrBlock: templates.Ref{"BOSHSubnetCIDR"},
					Tags: []templates.Tag{
						{
							Key:   "Name",
							Value: "BOSH",
						},
					},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHRouteTable", templates.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: templates.RouteTable{
					VpcId: templates.Ref{"VPC"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHRoute", templates.Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
				Properties: templates.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            templates.Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         templates.Ref{"BOSHRouteTable"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHSubnetRouteTableAssociation", templates.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: templates.SubnetRouteTableAssociation{
					RouteTableId: templates.Ref{"BOSHRouteTable"},
					SubnetId:     templates.Ref{"BOSHSubnet"},
				},
			}))

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("BOSHSubnetCIDR", templates.Parameter{
				Description: "CIDR block for the BOSH subnet.",
				Type:        "String",
				Default:     "10.0.0.0/24",
			}))
		})
	})

	Describe("InternalSubnet", func() {
		It("returns a template with all fields for the Internal subnet", func() {
			subnet := builder.InternalSubnet()

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("InternalSubnet", templates.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: templates.Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							"0",
							map[string]templates.Ref{
								"Fn::GetAZs": templates.Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: templates.Ref{"InternalSubnetCIDR"},
					VpcId:     templates.Ref{"VPC"},
					Tags: []templates.Tag{
						{
							Key:   "Name",
							Value: "Internal",
						},
					},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("InternalRouteTable", templates.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: templates.RouteTable{
					VpcId: templates.Ref{"VPC"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("InternalRoute", templates.Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "NATInstance",
				Properties: templates.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					RouteTableId:         templates.Ref{"InternalRouteTable"},
					InstanceId:           templates.Ref{"NATInstance"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("InternalSubnetRouteTableAssociation", templates.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: templates.SubnetRouteTableAssociation{
					RouteTableId: templates.Ref{"InternalRouteTable"},
					SubnetId:     templates.Ref{"InternalSubnet"},
				},
			}))

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("InternalSubnetCIDR", templates.Parameter{
				Description: "CIDR block for the Internal subnet.",
				Type:        "String",
				Default:     "10.0.16.0/20",
			}))
		})
	})

	Describe("LoadBalancerSubnet", func() {
		It("returns a template with all fields for the load balancer subnet", func() {
			subnet := builder.LoadBalancerSubnet()

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("LoadBalancerSubnetCIDR", templates.Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     "10.0.2.0/24",
			}))

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerSubnet", templates.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: templates.Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							"0",
							map[string]templates.Ref{
								"Fn::GetAZs": templates.Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: templates.Ref{"LoadBalancerSubnetCIDR"},
					VpcId:     templates.Ref{"VPC"},
					Tags: []templates.Tag{
						{
							Key:   "Name",
							Value: "LoadBalancer",
						},
					},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerRouteTable", templates.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: templates.RouteTable{
					VpcId: templates.Ref{"VPC"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerRoute", templates.Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
				Properties: templates.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            templates.Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         templates.Ref{"LoadBalancerRouteTable"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerSubnetRouteTableAssociation", templates.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: templates.SubnetRouteTableAssociation{
					RouteTableId: templates.Ref{"LoadBalancerRouteTable"},
					SubnetId:     templates.Ref{"LoadBalancerSubnet"},
				},
			}))
		})
	})
})
