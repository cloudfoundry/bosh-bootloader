package templates

type SecurityGroupTemplateBuilder struct{}

func NewSecurityGroupTemplateBuilder() SecurityGroupTemplateBuilder {
	return SecurityGroupTemplateBuilder{}
}

func (s SecurityGroupTemplateBuilder) ConcourseInternalSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"ConcourseInternalSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:            Ref{"VPC"},
					GroupDescription: "ConcourseInternal",
					SecurityGroupEgress: []SecurityGroupEgress{
						{
							SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							SourceSecurityGroupId: Ref{"ConcourseSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
				},
			},
		},
		Outputs: map[string]Output{
			"ConcourseInternalSecurityGroup": Output{
				Value: Ref{"ConcourseInternalSecurityGroup"},
			},
		},
	}
}

func (s SecurityGroupTemplateBuilder) CFRouterInternalSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"CFRouterInternalSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:            Ref{"VPC"},
					GroupDescription: "CFRouterInternal",
					SecurityGroupEgress: []SecurityGroupEgress{
						{
							SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							SourceSecurityGroupId: Ref{"CFRouterSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
				},
			},
		},
		Outputs: map[string]Output{
			"CFRouterInternalSecurityGroup": Output{
				Value: Ref{"CFRouterInternalSecurityGroup"},
			},
		},
	}
}

func (s SecurityGroupTemplateBuilder) CFSSHProxyInternalSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"CFSSHProxyInternalSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:            Ref{"VPC"},
					GroupDescription: "CFSSHProxyInternal",
					SecurityGroupEgress: []SecurityGroupEgress{
						{
							SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
					SecurityGroupIngress: []SecurityGroupIngress{
						{
							SourceSecurityGroupId: Ref{"CFSSHProxySecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
				},
			},
		},
		Outputs: map[string]Output{
			"CFSSHProxyInternalSecurityGroup": Output{
				Value: Ref{"CFSSHProxyInternalSecurityGroup"},
			},
		},
	}
}

func (s SecurityGroupTemplateBuilder) InternalSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"InternalSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Internal",
					SecurityGroupEgress: []SecurityGroupEgress{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress(nil, "tcp", "0", "65535", nil),
						s.securityGroupIngress(nil, "udp", "0", "65535", nil),
						s.securityGroupIngress("0.0.0.0/0", "icmp", "-1", "-1", nil),
					},
				},
			},
			"InternalSecurityGroupIngressTCPfromBOSH": s.internalSecurityGroupIngress("BOSHSecurityGroup", "tcp"),
			"InternalSecurityGroupIngressUDPfromBOSH": s.internalSecurityGroupIngress("BOSHSecurityGroup", "udp"),
			"InternalSecurityGroupIngressTCPfromSelf": s.internalSecurityGroupIngress("InternalSecurityGroup", "tcp"),
			"InternalSecurityGroupIngressUDPfromSelf": s.internalSecurityGroupIngress("InternalSecurityGroup", "udp"),
		},
		Outputs: map[string]Output{
			"InternalSecurityGroup": {Value: Ref{"InternalSecurityGroup"}},
		},
	}
}

func (s SecurityGroupTemplateBuilder) BOSHSecurityGroup() Template {
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
					SecurityGroupEgress: []SecurityGroupEgress{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress(Ref{"BOSHInboundCIDR"}, "tcp", "22", "22", nil),
						s.securityGroupIngress(Ref{"BOSHInboundCIDR"}, "tcp", "6868", "6868", nil),
						s.securityGroupIngress(Ref{"BOSHInboundCIDR"}, "tcp", "25555", "25555", nil),
						s.securityGroupIngress(nil, "tcp", "0", "65535", Ref{"InternalSecurityGroup"}),
						s.securityGroupIngress(nil, "udp", "0", "65535", Ref{"InternalSecurityGroup"}),
					},
				},
			},
		},
		Outputs: map[string]Output{
			"BOSHSecurityGroup": Output{Value: Ref{"BOSHSecurityGroup"}},
		},
	}
}

func (s SecurityGroupTemplateBuilder) ConcourseSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"ConcourseSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Concourse",
					SecurityGroupEgress: []SecurityGroupEgress{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress("0.0.0.0/0", "tcp", "80", "80", nil),
						s.securityGroupIngress("0.0.0.0/0", "tcp", "2222", "2222", nil),
					},
				},
			},
		},
	}
}

func (s SecurityGroupTemplateBuilder) CFRouterSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"CFRouterSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Router",
					SecurityGroupEgress: []SecurityGroupEgress{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress("0.0.0.0/0", "tcp", "80", "80", nil),
					},
				},
			},
		},
	}
}

func (s SecurityGroupTemplateBuilder) CFSSHProxySecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"CFSSHProxySecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "CFSSHProxy",
					SecurityGroupEgress: []SecurityGroupEgress{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress("0.0.0.0/0", "tcp", "2222", "2222", nil),
					},
				},
			},
		},
	}
}

func (SecurityGroupTemplateBuilder) internalSecurityGroupIngress(sourceSecurityGroupId, ipProtocol string) Resource {
	return Resource{
		Type: "AWS::EC2::SecurityGroupIngress",
		Properties: SecurityGroupIngress{
			GroupId:               Ref{"InternalSecurityGroup"},
			SourceSecurityGroupId: Ref{sourceSecurityGroupId},
			IpProtocol:            ipProtocol,
			FromPort:              "0",
			ToPort:                "65535",
		},
	}
}

func (SecurityGroupTemplateBuilder) securityGroupIngress(
	cidrIP interface{}, ipProtocol string, fromPort string, toPort string,
	sourceSecurityGroupId interface{}) SecurityGroupIngress {

	return SecurityGroupIngress{
		CidrIp:                cidrIP,
		IpProtocol:            ipProtocol,
		FromPort:              fromPort,
		ToPort:                toPort,
		SourceSecurityGroupId: sourceSecurityGroupId,
	}
}
