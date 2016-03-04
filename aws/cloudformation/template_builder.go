package cloudformation

type logger interface {
	Step(message string)
	Dot()
}

type TemplateBuilder struct {
	logger logger
}

func NewTemplateBuilder(logger logger) TemplateBuilder {
	return TemplateBuilder{
		logger: logger,
	}
}

func (t TemplateBuilder) Build(keyPairName string) Template {
	t.logger.Step("generating cloudformation template")

	return Template{
		AWSTemplateFormatVersion: "2010-09-09",
		Description:              "Infrastructure for a MicroBOSH deployment with an ELB.",
	}.Merge(t.SSHKeyPairName(keyPairName),
		t.BOSHIAMUser(),
		t.NAT(),
		t.VPC(),
		t.BOSHSubnet(),
		t.InternalSubnet(),
		t.LoadBalancerSubnet(),
		t.InternalSecurityGroup(),
		t.BOSHSecurityGroup(),
		t.WebSecurityGroup(),
		t.WebELBLoadBalancer(),
		t.MicroEIP(),
	)
}

func (t TemplateBuilder) BOSHIAMUser() Template {
	return Template{
		Resources: map[string]Resource{
			"BOSHUser": Resource{
				Type: "AWS::IAM::User",
				Properties: IAMUser{
					Policies: []IAMPolicy{
						{
							PolicyName: "aws-cpi",
							PolicyDocument: IAMPolicyDocument{
								Version: "2012-10-17",
								Statement: []IAMStatement{
									{
										Action: []string{
											"ec2:AssociateAddress",
											"ec2:AttachVolume",
											"ec2:CreateVolume",
											"ec2:DeleteSnapshot",
											"ec2:DeleteVolume",
											"ec2:DescribeAddresses",
											"ec2:DescribeImages",
											"ec2:DescribeInstances",
											"ec2:DescribeRegions",
											"ec2:DescribeSecurityGroups",
											"ec2:DescribeSnapshots",
											"ec2:DescribeSubnets",
											"ec2:DescribeVolumes",
											"ec2:DetachVolume",
											"ec2:CreateSnapshot",
											"ec2:CreateTags",
											"ec2:RunInstances",
											"ec2:TerminateInstances",
										},
										Effect:   "Allow",
										Resource: "*",
									},
									{
										Action:   []string{"elasticloadbalancing:*"},
										Effect:   "Allow",
										Resource: "*",
									},
								},
							},
						},
					},
				},
			},
			"BOSHUserAccessKey": Resource{
				Properties: IAMAccessKey{
					UserName: Ref{"BOSHUser"},
				},
				Type: "AWS::IAM::AccessKey",
			},
		},
		Outputs: map[string]Output{
			"BOSHUserAccessKey": Output{
				Value: Ref{"BOSHUserAccessKey"},
			},
			"BOSHUserSecretAccessKey": Output{
				Value: FnGetAtt{
					[]string{
						"BOSHUserAccessKey",
						"SecretAccessKey",
					},
				},
			},
		},
	}
}

func (t TemplateBuilder) NAT() Template {
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
					SecurityGroupEgress: []string{},
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
					InstanceType:    "m4.large",
					SubnetId:        Ref{"BOSHSubnet"},
					SourceDestCheck: false,
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
				Type: "AWS::EC2::EIP",
				Properties: EIP{
					Domain:     "vpc",
					InstanceId: Ref{"NATInstance"},
				},
			},
		},
	}
}

func (t TemplateBuilder) VPC() Template {
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

func (t TemplateBuilder) BOSHSubnet() Template {
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
					VpcId:     Ref{"VPC"},
					CidrBlock: Ref{"BOSHSubnetCIDR"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "BOSH",
						},
					},
				},
			},
			"BOSHRouteTable": Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
				},
			},
			"BOSHRoute": Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
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
	}
}

func (t TemplateBuilder) InternalSubnet() Template {
	return Template{
		Parameters: map[string]Parameter{
			"InternalSubnetCIDR": Parameter{
				Description: "CIDR block for the Internal subnet.",
				Type:        "String",
				Default:     "10.0.16.0/20",
			},
		},
		Resources: map[string]Resource{
			"InternalSubnet": Resource{
				Type: "AWS::EC2::Subnet",
				Properties: Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							"0",
							map[string]Ref{
								"Fn::GetAZs": Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: Ref{"InternalSubnetCIDR"},
					VpcId:     Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "Internal",
						},
					},
				},
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
			"InternalSubnetRouteTableAssociation": Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"InternalRouteTable"},
					SubnetId:     Ref{"InternalSubnet"},
				},
			},
		},
	}
}

func (t TemplateBuilder) LoadBalancerSubnet() Template {
	return Template{
		Parameters: map[string]Parameter{
			"LoadBalancerSubnetCIDR": Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     "10.0.2.0/24",
			},
		},
		Resources: map[string]Resource{
			"LoadBalancerSubnet": Resource{
				Type: "AWS::EC2::Subnet",
				Properties: Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							"0",
							map[string]Ref{
								"Fn::GetAZs": Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: Ref{"LoadBalancerSubnetCIDR"},
					VpcId:     Ref{"VPC"},
					Tags: []Tag{
						{
							Key:   "Name",
							Value: "LoadBalancer",
						},
					},
				},
			},
			"LoadBalancerRouteTable": Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: RouteTable{
					VpcId: Ref{"VPC"},
				},
			},
			"LoadBalancerRoute": Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
				Properties: Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         Ref{"LoadBalancerRouteTable"},
				},
			},
			"LoadBalancerSubnetRouteTableAssociation": {
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: SubnetRouteTableAssociation{
					RouteTableId: Ref{"LoadBalancerRouteTable"},
					SubnetId:     Ref{"LoadBalancerSubnet"},
				},
			},
		},
	}
}

func (t TemplateBuilder) InternalSecurityGroup() Template {
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
							SourceSecurityGroupId: Ref{"WebSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
						{
							SourceSecurityGroupId: Ref{"WebSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
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

func (t TemplateBuilder) BOSHSecurityGroup() Template {
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
	}
}

func (t TemplateBuilder) WebSecurityGroup() Template {
	return Template{
		Resources: map[string]Resource{
			"WebSecurityGroup": Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: SecurityGroup{
					VpcId:               Ref{"VPC"},
					GroupDescription:    "Web",
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
		},
	}
}

func (t TemplateBuilder) WebELBLoadBalancer() Template {
	return Template{
		Resources: map[string]Resource{
			"WebELBLoadBalancer": {
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: ElasticLoadBalancingLoadBalancer{
					Subnets:        []interface{}{Ref{"LoadBalancerSubnet"}},
					SecurityGroups: []interface{}{Ref{"WebSecurityGroup"}},

					HealthCheck: HealthCheck{
						HealthyThreshold:   "2",
						Interval:           "30",
						Target:             "tcp:8080",
						Timeout:            "5",
						UnhealthyThreshold: "10",
					},

					Listeners: []Listener{
						{
							Protocol:         "tcp",
							LoadBalancerPort: "80",
							InstanceProtocol: "tcp",
							InstancePort:     "8080",
						},
						{
							Protocol:         "tcp",
							LoadBalancerPort: "2222",
							InstanceProtocol: "tcp",
							InstancePort:     "2222",
						},
					},
				},
			},
		},
	}
}

func (t TemplateBuilder) SSHKeyPairName(keyPairName string) Template {
	return Template{
		Parameters: map[string]Parameter{
			"SSHKeyPairName": Parameter{
				Type:        "AWS::EC2::KeyPair::KeyName",
				Default:     keyPairName,
				Description: "SSH KeyPair to use for instances",
			},
		},
	}
}

func (t TemplateBuilder) MicroEIP() Template {
	return Template{
		Resources: map[string]Resource{
			"MicroEIP": Resource{
				Type: "AWS::EC2::EIP",
				Properties: EIP{
					Domain: "vpc",
				},
			},
		},
	}
}
