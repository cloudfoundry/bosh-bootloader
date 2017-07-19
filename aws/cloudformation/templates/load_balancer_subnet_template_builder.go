package templates

import "fmt"

type LoadBalancerSubnetTemplateBuilder struct{}

func NewLoadBalancerSubnetTemplateBuilder() LoadBalancerSubnetTemplateBuilder {
	return LoadBalancerSubnetTemplateBuilder{}
}

func (LoadBalancerSubnetTemplateBuilder) LoadBalancerSubnet(az, subnetSuffix, cidrBlock string) Template {
	subnetName := fmt.Sprintf("LoadBalancerSubnet%s", subnetSuffix)
	subnetID := fmt.Sprintf("%sName", subnetName)
	cidrName := fmt.Sprintf("%sCIDR", subnetName)
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
					AvailabilityZone: az,
					CidrBlock:        Ref{cidrName},
					VpcId:            Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: tag,
						},
					},
				},
				DeletionPolicy: "Retain",
			},
			"LoadBalancerRouteTable": Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
				},
				DeletionPolicy: "Retain",
			},
			"LoadBalancerRoute": Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayAttachment",
				Properties: Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         Ref{"LoadBalancerRouteTable"},
				},
				DeletionPolicy: "Retain",
			},
			routeTableAssociationName: {
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"LoadBalancerRouteTable"},
					SubnetId:     Ref{subnetName},
				},
				DeletionPolicy: "Retain",
			},
		},
		Outputs: map[string]Output{
			subnetID: Output{
				Value: Ref{
					Ref: subnetName,
				},
			},
			"LoadBalancerRouteTable": Output{
				Value: Ref{"LoadBalancerRouteTable"},
			},
		},
	}
}
