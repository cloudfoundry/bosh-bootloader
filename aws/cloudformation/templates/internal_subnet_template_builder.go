package templates

import "fmt"

type InternalSubnetTemplateBuilder struct{}

func NewInternalSubnetTemplateBuilder() InternalSubnetTemplateBuilder {
	return InternalSubnetTemplateBuilder{}
}

func (s InternalSubnetTemplateBuilder) InternalSubnet(azIndex int, suffix, cidrBlock string, envID string) Template {
	subnetName := fmt.Sprintf("InternalSubnet%s", suffix)
	subnetTag := fmt.Sprintf("Internal%s", suffix)
	subnetCIDRName := fmt.Sprintf("%sCIDR", subnetName)
	cidrDescription := fmt.Sprintf("CIDR block for %s.", subnetName)
	subnetRouteTableAssociationName := fmt.Sprintf("%sRouteTableAssociation", subnetName)

	return Template{
		Outputs: map[string]Output{
			fmt.Sprintf("%sName", subnetName): Output{
				Value: Ref{subnetName},
			},
			fmt.Sprintf("%sAZ", subnetName): Output{
				FnGetAtt{
					[]string{
						subnetName,
						"AvailabilityZone",
					},
				},
			},
			subnetCIDRName: Output{
				Value: Ref{subnetCIDRName},
			},
		},
		Parameters: map[string]Parameter{
			subnetCIDRName: Parameter{
				Description: cidrDescription,
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
							fmt.Sprintf("%d", azIndex),
							map[string]Ref{
								"Fn::GetAZs": Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: Ref{subnetCIDRName},
					VpcId:     Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: subnetTag,
						},
						{
							Key:   "bbl-env-id",
							Value: envID,
						},
					},
				},
			},
			"InternalRouteTable": {
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "bbl-env-id",
							Value: envID,
						},
					},
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
			subnetRouteTableAssociationName: Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"InternalRouteTable"},
					SubnetId:     Ref{subnetName},
				},
			},
		},
	}
}
