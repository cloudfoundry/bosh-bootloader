package templates

type NATTemplateBuilder struct{}

func NewNATTemplateBuilder() NATTemplateBuilder {
	return NATTemplateBuilder{}
}

func (t NATTemplateBuilder) NAT() Template {
	return Template{
		Mappings: map[string]interface{}{
			"AWSNATAMI": map[string]AMI{
				"us-east-1":      {"ami-68115b02"},
				"us-west-1":      {"ami-ef1a718f"},
				"us-west-2":      {"ami-77a4b816"},
				"eu-west-1":      {"ami-c0993ab3"},
				"eu-central-1":   {"ami-0b322e67"},
				"ap-southeast-1": {"ami-e2fc3f81"},
				"ap-southeast-2": {"ami-e3217a80"},
				"ap-northeast-1": {"ami-f885ae96"},
				"ap-northeast-2": {"ami-4118d72f"},
				"sa-east-1":      {"ami-8631b5ea"},
			},
		},
		Resources: map[string]Resource{
			"NATSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "NAT",
					SecurityGroupEgress: []SecurityGroupEgress{},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
						{
							SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
				},
			},
			"NATInstance": Resource{
				Type: "AWS::EC2::Instance",
				Properties: Instance{
					PrivateIpAddress: "10.0.0.7",
					InstanceType:     "t2.medium",
					SubnetId:         Ref{"BOSHSubnet"},
					SourceDestCheck:  false,
					ImageId: map[string]interface{}{
						"Fn::FindInMap": []interface{}{
							"AWSNATAMI",
							Ref{"AWS::Region"},
							"AMI",
						},
					},
					KeyName: Ref{"SSHKeyPairName"},
					SecurityGroupIds: []interface{}{
						Ref{"NATSecurityGroup"},
					},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "NAT",
						},
					},
				},
			},
			"NATEIP": Resource{
				DependsOn: "VPCGatewayAttachment",
				Type:      "AWS::EC2::EIP",
				Properties: EIP{
					Domain:     "vpc",
					InstanceId: Ref{"NATInstance"},
				},
			},
		},
	}
}
