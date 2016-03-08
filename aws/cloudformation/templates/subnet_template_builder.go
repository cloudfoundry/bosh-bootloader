package templates

type SubnetTemplateBuilder struct {
}

func NewSubnetTemplateBuilder() SubnetTemplateBuilder {
	return SubnetTemplateBuilder{}
}

func (s SubnetTemplateBuilder) BOSHSubnet() Template {
	return Template{
		Parameters: map[string]Parameter{
			"BOSHSubnetCIDR": Parameter{
				Description: "CIDR block for the BOSH subnet.",
				Type:        "String",
				Default:     "10.0.0.0/24",
			},
		},
		Resources: map[string]Resource{
			"BOSHSubnet": Resource{
				Type: "AWS::EC2::Subnet",
				Properties: Subnet{
					VpcId:     Ref{"VPC"},
					CidrBlock: Ref{"BOSHSubnetCIDR"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "BOSH",
						},
					},
				},
			},
			"BOSHRouteTable": Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
				},
			},
			"BOSHRoute": Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
				Properties: Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         Ref{"BOSHRouteTable"},
				},
			},
			"BOSHSubnetRouteTableAssociation": Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"BOSHRouteTable"},
					SubnetId:     Ref{"BOSHSubnet"},
				},
			},
		},
		Outputs: map[string]Output{
			"BOSHSubnet": Output{
				Value: Ref{"BOSHSubnet"},
			},
			"BOSHSubnetAZ": Output{
				Value: FnGetAtt{
					[]string{
						"BOSHSubnet",
						"AvailabilityZone",
					},
				},
			},
		},
	}
}

func (s SubnetTemplateBuilder) InternalSubnet() Template {
	return Template{
		Parameters: map[string]Parameter{
			"InternalSubnetCIDR": Parameter{
				Description: "CIDR block for the Internal subnet.",
				Type:        "String",
				Default:     "10.0.16.0/20",
			},
		},
		Resources: map[string]Resource{
			"InternalSubnet": Resource{
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
					CidrBlock: Ref{"InternalSubnetCIDR"},
					VpcId:     Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "Internal",
						},
					},
				},
			},
			"InternalRouteTable": {
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
				},
			},
			"InternalRoute": {
				Type:      "AWS::EC2::Route",
				DependsOn: "NATInstance",
				Properties: Route{
					DestinationCidrBlock: "0.0.0.0/0",
					RouteTableId:         Ref{"InternalRouteTable"},
					InstanceId:           Ref{"NATInstance"},
				},
			},
			"InternalSubnetRouteTableAssociation": Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"InternalRouteTable"},
					SubnetId:     Ref{"InternalSubnet"},
				},
			},
		},
	}
}

func (s SubnetTemplateBuilder) LoadBalancerSubnet() Template {
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
