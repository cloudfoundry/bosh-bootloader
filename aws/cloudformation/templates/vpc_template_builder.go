package templates

import "fmt"

type VPCTemplateBuilder struct{}

func NewVPCTemplateBuilder() VPCTemplateBuilder {
	return VPCTemplateBuilder{}
}

func (t VPCTemplateBuilder) VPC(envID string) Template {
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
							Value: fmt.Sprintf("vpc-%s", envID),
							Key:   "Name",
						},
					},
				},
				DeletionPolicy: "Retain",
			},
			"VPCGatewayInternetGateway": Resource{
				Type:           "AWS::EC2::InternetGateway",
				DeletionPolicy: "Retain",
			},
			"VPCGatewayAttachment": Resource{
				Type: "AWS::EC2::VPCGatewayAttachment",
				Properties: VPCGatewayAttachment{
					VpcId:             Ref{"VPC"},
					InternetGatewayId: Ref{"VPCGatewayInternetGateway"},
				},
				DeletionPolicy: "Retain",
			},
		},

		Outputs: map[string]Output{
			"VPCID": Output{
				Value: Ref{
					Ref: "VPC",
				},
			},
			"VPCInternetGatewayID": Output{
				Value: Ref{
					Ref: "VPCGatewayInternetGateway",
				},
			},
		},
	}
}
