package templates

import "fmt"

type LoadBalancerSubnetTemplateBuilder struct{}

func NewLoadBalancerSubnetTemplateBuilder() LoadBalancerSubnetTemplateBuilder {
	return LoadBalancerSubnetTemplateBuilder{}
}

func (LoadBalancerSubnetTemplateBuilder) LoadBalancerSubnet(azIndex int, subnetSuffix string, cidrBlock string) Template {
	return Template{
		Parameters: map[string]Parameter{
			fmt.Sprintf("LoadBalancerSubnet%sCIDR", subnetSuffix): Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     cidrBlock,
			},
		},
		Resources: map[string]Resource{
			fmt.Sprintf("LoadBalancerSubnet%s", subnetSuffix): Resource{
				Type: "AWS::EC2::Subnet",
				Properties: Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							fmt.Sprintf("%d", azIndex),
							map[string]Ref{
								"Fn::GetAZs": Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: Ref{fmt.Sprintf("LoadBalancerSubnet%sCIDR", subnetSuffix)},
					VpcId:     Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: fmt.Sprintf("LoadBalancer%s", subnetSuffix),
						},
					},
				},
			},
			"LoadBalancerRouteTable": Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
				},
			},
			"LoadBalancerRoute": Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
				Properties: Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         Ref{"LoadBalancerRouteTable"},
				},
			},
			fmt.Sprintf("LoadBalancerSubnet%sRouteTableAssociation", subnetSuffix): {
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"LoadBalancerRouteTable"},
					SubnetId:     Ref{fmt.Sprintf("LoadBalancerSubnet%s", subnetSuffix)},
				},
			},
		},
	}
}
