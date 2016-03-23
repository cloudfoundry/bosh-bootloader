package templates

type LoadBalancerSubnetTemplateBuilder struct{}

func NewLoadBalancerSubnetTemplateBuilder() LoadBalancerSubnetTemplateBuilder {
	return LoadBalancerSubnetTemplateBuilder{}
}

func (LoadBalancerSubnetTemplateBuilder) LoadBalancerSubnet() Template {
	return Template{
		Parameters: map[string]Parameter{
			"LoadBalancerSubnetCIDR": Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     "10.0.2.0/24",
			},
		},
		Resources: map[string]Resource{
			"LoadBalancerSubnet": Resource{
				Type: "AWS::EC2::Subnet",
				Properties: Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							"0",
							map[string]Ref{
								"Fn::GetAZs": Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: Ref{"LoadBalancerSubnetCIDR"},
					VpcId:     Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "LoadBalancer",
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
			"LoadBalancerSubnetRouteTableAssociation": {
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"LoadBalancerRouteTable"},
					SubnetId:     Ref{"LoadBalancerSubnet"},
				},
			},
		},
	}
}
