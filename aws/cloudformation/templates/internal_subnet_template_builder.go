package templates

import "fmt"

type InternalSubnetTemplateBuilder struct{}

func NewInternalSubnetTemplateBuilder() InternalSubnetTemplateBuilder {
	return InternalSubnetTemplateBuilder{}
}

func (s InternalSubnetTemplateBuilder) InternalSubnet(az, suffix, cidrBlock string) Template {
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
					AvailabilityZone: az,
					CidrBlock:        Ref{subnetCIDRName},
					VpcId:            Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: subnetTag,
						},
					},
				},
				DeletionPolicy: "Retain",
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
