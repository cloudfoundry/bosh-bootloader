package templates

type BOSHSubnetTemplateBuilder struct{}

func NewBOSHSubnetTemplateBuilder() BOSHSubnetTemplateBuilder {
	return BOSHSubnetTemplateBuilder{}
}

func (BOSHSubnetTemplateBuilder) BOSHSubnet(availabilityZone string) Template {
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
					VpcId:            Ref{"VPC"},
					CidrBlock:        Ref{"BOSHSubnetCIDR"},
					AvailabilityZone: availabilityZone,
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "BOSH",
						},
					},
				},
				DeletionPolicy: "Retain",
			},
			"BOSHRouteTable": Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
				},
			},
			"BOSHRoute": Resource{
				DependsOn: "VPCGatewayAttachment",
				Type:      "AWS::EC2::Route",
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
		},
	}
}
