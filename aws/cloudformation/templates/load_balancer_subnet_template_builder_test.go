package templates_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
)

var _ = Describe("LoadBalancerSubnetTemplateBuilder", func() {
	var builder templates.LoadBalancerSubnetTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewLoadBalancerSubnetTemplateBuilder()
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
