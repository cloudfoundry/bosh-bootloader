package templates

type SecurityGroupTemplateBuilder struct{}

func NewSecurityGroupTemplateBuilder() SecurityGroupTemplateBuilder {
	return SecurityGroupTemplateBuilder{}
}

func (s SecurityGroupTemplateBuilder) InternalSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"InternalSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Internal",
					SecurityGroupEgress: []string{},
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
					SecurityGroupEgress: []string{},
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
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress("0.0.0.0/0", "tcp", "80", "80", nil),
						s.securityGroupIngress("0.0.0.0/0", "tcp", "2222", "2222", nil),
						s.securityGroupIngress("0.0.0.0/0", "tcp", "443", "443", nil),
					},
				},
			},
			"InternalSecurityGroupIngressTCPfromConcourseSecurityGroup": s.internalSecurityGroupIngress("ConcourseSecurityGroup", "tcp"),
			"InternalSecurityGroupIngressUDPfromConcourseSecurityGroup": s.internalSecurityGroupIngress("ConcourseSecurityGroup", "udp"),
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
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress("0.0.0.0/0", "tcp", "80", "80", nil),
						s.securityGroupIngress("0.0.0.0/0", "tcp", "2222", "2222", nil),
						s.securityGroupIngress("0.0.0.0/0", "tcp", "443", "443", nil),
					},
				},
			},
			"InternalSecurityGroupIngressTCPfromCFRouterSecurityGroup": s.internalSecurityGroupIngress("CFRouterSecurityGroup", "tcp"),
			"InternalSecurityGroupIngressUDPfromCFRouterSecurityGroup": s.internalSecurityGroupIngress("CFRouterSecurityGroup", "udp"),
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
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []SecurityGroupIngress{
						s.securityGroupIngress("0.0.0.0/0", "tcp", "2222", "2222", nil),
					},
				},
			},
			"InternalSecurityGroupIngressTCPfromCFSSHProxySecurityGroup": s.internalSecurityGroupIngress("CFSSHProxySecurityGroup", "tcp"),
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
