package templates

type VPCTemplateBuilder struct{}

func NewVPCTemplateBuilder() VPCTemplateBuilder {
	return VPCTemplateBuilder{}
}

func (t VPCTemplateBuilder) VPC() Template {
	return Template{
		Parameters: map[string]Parameter{
			"VPCCIDR": Parameter{
				Description: "CIDR block for the VPC.",
				Type:        "String",
				Default:     "10.0.0.0/16",
			},
		},
		Resources: map[string]Resource{
			"VPC": Resource{
				Type: "AWS::EC2::VPC",
				Properties: VPC{
					CidrBlock: Ref{"VPCCIDR"},
					Tags: []Tag{
						{
							Value: "concourse",
							Key:   "Name",
						},
					},
				},
			},
			"VPCGatewayInternetGateway": Resource{
				Type: "AWS::EC2::InternetGateway",
			},
			"VPCGatewayAttachment": Resource{
				Type: "AWS::EC2::VPCGatewayAttachment",
				Properties: VPCGatewayAttachment{
					VpcId:             Ref{"VPC"},
					InternetGatewayId: Ref{"VPCGatewayInternetGateway"},
				},
			},
		},
	}
}
