package templates

import "fmt"

type LoadBalancerSubnetTemplateBuilder struct{}

func NewLoadBalancerSubnetTemplateBuilder() LoadBalancerSubnetTemplateBuilder {
	return LoadBalancerSubnetTemplateBuilder{}
}

func (LoadBalancerSubnetTemplateBuilder) LoadBalancerSubnet(azIndex int, subnetSuffix, cidrBlock, envID string) Template {
	subnetName := fmt.Sprintf("LoadBalancerSubnet%s", subnetSuffix)
	cidrName := fmt.Sprintf("%sCIDR", subnetName)
	az := fmt.Sprintf("%d", azIndex)
	tag := fmt.Sprintf("LoadBalancer%s", subnetSuffix)
	routeTableAssociationName := fmt.Sprintf("%sRouteTableAssociation", subnetName)

	return Template{
		Parameters: map[string]Parameter{
			cidrName: Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     cidrBlock,
			},
		},
		Resources: map[string]Resource{
			subnetName: Resource{
				Type: "AWS::EC2::Subnet",
				Properties: Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							az,
							map[string]Ref{
								"Fn::GetAZs": Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: Ref{cidrName},
					VpcId:     Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: tag,
						},
						{
							Key:   bblTagKey,
							Value: envID,
						},
					},
				},
			},
			"LoadBalancerRouteTable": Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   bblTagKey,
							Value: envID,
						},
					},
				},
			},
			"LoadBalancerRoute": Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayAttachment",
				Properties: Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         Ref{"LoadBalancerRouteTable"},
				},
			},
			routeTableAssociationName: {
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"LoadBalancerRouteTable"},
					SubnetId:     Ref{subnetName},
				},
			},
		},
	}
}
