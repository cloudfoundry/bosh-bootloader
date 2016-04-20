package templates

type SecurityGroupTemplateBuilder struct{}

func NewSecurityGroupTemplateBuilder() SecurityGroupTemplateBuilder {
	return SecurityGroupTemplateBuilder{}
}

func (t SecurityGroupTemplateBuilder) InternalSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"InternalSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Internal",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							IpProtocol: "tcp",
							FromPort:   "0",
							ToPort:     "65535",
						},
						{
							IpProtocol: "udp",
							FromPort:   "0",
							ToPort:     "65535",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "icmp",
							FromPort:   "-1",
							ToPort:     "-1",
						},
					},
				},
			},
			"InternalSecurityGroupIngressTCPfromBOSH": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"BOSHSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
			"InternalSecurityGroupIngressUDPfromBOSH": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"BOSHSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
			"InternalSecurityGroupIngressTCPfromSelf": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
			"InternalSecurityGroupIngressUDPfromSelf": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
		},
	}
}

func (t SecurityGroupTemplateBuilder) BOSHSecurityGroup() Template {
	return Template{
		Parameters: map[string]Parameter{
			"BOSHInboundCIDR": Parameter{
				Description: "CIDR to permit access to BOSH (e.g. 205.103.216.37/32 for your specific IP)",
				Type:        "String",
				Default:     "0.0.0.0/0",
			},
		},
		Resources: map[string]Resource{
			"BOSHSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "BOSH",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							CidrIp:     Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "22",
							ToPort:     "22",
						},

						{
							CidrIp:     Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "6868",
							ToPort:     "6868",
						},
						{
							CidrIp:     Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "25555",
							ToPort:     "25555",
						},
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
		},
		Outputs: map[string]Output{
			"BOSHSecurityGroup": Output{Value: Ref{"BOSHSecurityGroup"}},
		},
	}
}

func (t SecurityGroupTemplateBuilder) ConcourseSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"ConcourseSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Concourse",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "80",
							ToPort:     "80",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "2222",
							ToPort:     "2222",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "443",
							ToPort:     "443",
						},
					},
				},
			},
			"InternalSecurityGroupIngressTCPfromConcourseSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"ConcourseSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
			"InternalSecurityGroupIngressUDPfromConcourseSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"ConcourseSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
		},
	}
}

func (SecurityGroupTemplateBuilder) RouterSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"RouterSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Router",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "80",
							ToPort:     "80",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "2222",
							ToPort:     "2222",
						},
						{
							CidrIp:     "0.0.0.0/0",
							IpProtocol: "tcp",
							FromPort:   "443",
							ToPort:     "443",
						},
					},
				},
			},
			"InternalSecurityGroupIngressTCPfromRouterSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"RouterSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
			"InternalSecurityGroupIngressUDPfromRouterSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: SecurityGroupIngress{
					GroupId:               Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: Ref{"RouterSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			},
		},
	}
}
