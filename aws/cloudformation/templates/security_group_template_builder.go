package templates

type SecurityGroupTemplateBuilder struct{}

func NewSecurityGroupTemplateBuilder() SecurityGroupTemplateBuilder {
	return SecurityGroupTemplateBuilder{}
}

func (s SecurityGroupTemplateBuilder) LBSecurityGroup(securityGroupName, securityGroupDescription,
	loadBalancerName string, template Template) Template {
	securityGroupIngress := []SecurityGroupIngress{}

	properties := template.Resources[loadBalancerName].Properties.(ElasticLoadBalancingLoadBalancer)

	for _, listener := range properties.Listeners {
		securityGroupIngress = append(securityGroupIngress, s.securityGroupIngress(
			"0.0.0.0/0",
			s.determineSecurityGroupProtocol(listener.Protocol),
			listener.LoadBalancerPort,
			listener.LoadBalancerPort,
			nil,
		))
	}

	return Template{
		Resources: map[string]Resource{
			securityGroupName: Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:                Ref{"VPC"},
					GroupDescription:     securityGroupDescription,
					SecurityGroupEgress:  []SecurityGroupEgress{},
					SecurityGroupIngress: securityGroupIngress,
				},
			},
		},
	}
}

func (s SecurityGroupTemplateBuilder) LBInternalSecurityGroup(securityGroupName, lbSecurityGroupName,
	securityGroupDescription, loadBalancerName string, template Template) Template {

	securityGroupEgress := []SecurityGroupEgress{}
	securityGroupIngress := []SecurityGroupIngress{}
	securityGroupPorts := map[string]bool{}

	properties := template.Resources[loadBalancerName].Properties.(ElasticLoadBalancingLoadBalancer)
	for _, listener := range properties.Listeners {
		if !securityGroupPorts[listener.InstancePort] {
			securityGroupIngress = append(securityGroupIngress, SecurityGroupIngress{
				SourceSecurityGroupId: Ref{lbSecurityGroupName},
				IpProtocol:            s.determineSecurityGroupProtocol(listener.Protocol),
				FromPort:              listener.InstancePort,
			})
			securityGroupEgress = append(securityGroupEgress, SecurityGroupEgress{
				SourceSecurityGroupId: Ref{"InternalSecurityGroup"},
				IpProtocol:            s.determineSecurityGroupProtocol(listener.Protocol),
				FromPort:              listener.InstancePort,
			})

			securityGroupPorts[listener.InstancePort] = true
		}
	}

	return Template{
		Resources: map[string]Resource{
			securityGroupName: Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:                Ref{"VPC"},
					GroupDescription:     securityGroupDescription,
					SecurityGroupEgress:  securityGroupEgress,
					SecurityGroupIngress: securityGroupIngress,
				},
			},
		},
		Outputs: map[string]Output{
			securityGroupName: Output{
				Value: Ref{securityGroupName},
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

func (SecurityGroupTemplateBuilder) determineSecurityGroupProtocol(listenerProtocol string) string {
	switch listenerProtocol {
	case "ssl":
		return "tcp"
	case "http", "https":
		return "tcp"
	default:
		return listenerProtocol
	}
}
