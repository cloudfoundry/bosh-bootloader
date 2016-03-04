package cloudformation_test

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateBuilder", func() {
	var (
		builder cloudformation.TemplateBuilder
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		builder = cloudformation.NewTemplateBuilder(logger)
	})

	Describe("Build", func() {
		It("builds a cloudformation template", func() {
			template := builder.Build("keypair-name")
			Expect(template.Parameters).To(HaveKey("SSHKeyPairName"))
			Expect(template.Resources).To(HaveKey("BOSHUser"))
			Expect(template.Resources).To(HaveKey("NATInstance"))
			Expect(template.Resources).To(HaveKey("VPC"))
			Expect(template.Resources).To(HaveKey("BOSHSubnet"))
			Expect(template.Resources).To(HaveKey("InternalSubnet"))
			Expect(template.Resources).To(HaveKey("LoadBalancerSubnet"))
			Expect(template.Resources).To(HaveKey("InternalSecurityGroup"))
			Expect(template.Resources).To(HaveKey("BOSHSecurityGroup"))
			Expect(template.Resources).To(HaveKey("WebSecurityGroup"))
			Expect(template.Resources).To(HaveKey("WebELBLoadBalancer"))
			Expect(template.Resources).To(HaveKey("MicroEIP"))
		})

		It("logs that the cloudformation template is being generated", func() {
			builder.Build("keypair-name")

			Expect(logger.StepCall.Receives.Message).To(Equal("generating cloudformation template"))
		})
	})

	Describe("BOSHIAMUser", func() {
		It("returns a template for a BOSH IAM user", func() {
			user := builder.BOSHIAMUser()
			Expect(user.Resources).To(HaveLen(2))
			Expect(user.Resources).To(HaveKeyWithValue("BOSHUser", cloudformation.Resource{
				Type: "AWS::IAM::User",
				Properties: cloudformation.IAMUser{
					Policies: []cloudformation.IAMPolicy{
						{
							PolicyName: "aws-cpi",
							PolicyDocument: cloudformation.IAMPolicyDocument{
								Version: "2012-10-17",
								Statement: []cloudformation.IAMStatement{
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
			}))

			Expect(user.Resources).To(HaveKeyWithValue("BOSHUserAccessKey", cloudformation.Resource{
				Properties: cloudformation.IAMAccessKey{
					UserName: cloudformation.Ref{"BOSHUser"},
				},
				Type: "AWS::IAM::AccessKey",
			}))

			Expect(user.Outputs).To(HaveLen(2))
			Expect(user.Outputs).To(HaveKeyWithValue("BOSHUserAccessKey", cloudformation.Output{
				Value: cloudformation.Ref{"BOSHUserAccessKey"},
			}))

			Expect(user.Outputs).To(HaveKeyWithValue("BOSHUserSecretAccessKey", cloudformation.Output{
				Value: cloudformation.FnGetAtt{
					[]string{
						"BOSHUserAccessKey",
						"SecretAccessKey",
					},
				},
			}))
		})
	})

	Describe("NAT", func() {
		It("returns a template containing all of the NAT fields", func() {
			nat := builder.NAT()

			Expect(nat.Mappings).To(HaveLen(1))
			Expect(nat.Mappings).To(HaveKeyWithValue("AWSNATAMI", map[string]cloudformation.AMI{
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
			}))

			Expect(nat.Resources).To(HaveLen(3))
			Expect(nat.Resources).To(HaveKeyWithValue("NATSecurityGroup", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: cloudformation.SecurityGroup{
					VpcId:               cloudformation.Ref{"VPC"},
					GroupDescription:    "NAT",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []cloudformation.SecurityGroupIngress{
						{
							SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
						{
							SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
				},
			}))

			Expect(nat.Resources).To(HaveKeyWithValue("NATInstance", cloudformation.Resource{
				Type: "AWS::EC2::Instance",
				Properties: cloudformation.Instance{
					InstanceType:    "m4.large",
					SubnetId:        cloudformation.Ref{"BOSHSubnet"},
					SourceDestCheck: false,
					ImageId: map[string]interface{}{
						"Fn::FindInMap": []interface{}{
							"AWSNATAMI",
							cloudformation.Ref{"AWS::Region"},
							"AMI",
						},
					},
					KeyName: cloudformation.Ref{"SSHKeyPairName"},
					SecurityGroupIds: []interface{}{
						cloudformation.Ref{"NATSecurityGroup"},
					},
					Tags: []cloudformation.Tag{
						{
							Key:   "Name",
							Value: "NAT",
						},
					},
				},
			}))
			Expect(nat.Resources).To(HaveKeyWithValue("NATEIP", cloudformation.Resource{
				Type: "AWS::EC2::EIP",
				Properties: cloudformation.EIP{
					Domain:     "vpc",
					InstanceId: cloudformation.Ref{"NATInstance"},
				},
			}))
		})
	})

	Describe("VPC", func() {
		It("returns a template with all the VPC-related fields", func() {
			vpc := builder.VPC()

			Expect(vpc.Resources).To(HaveLen(3))
			Expect(vpc.Resources).To(HaveKeyWithValue("VPC", cloudformation.Resource{
				Type: "AWS::EC2::VPC",
				Properties: cloudformation.VPC{
					CidrBlock: cloudformation.Ref{"VPCCIDR"},
					Tags: []cloudformation.Tag{
						{
							Value: "concourse",
							Key:   "Name",
						},
					},
				},
			}))

			Expect(vpc.Resources).To(HaveKeyWithValue("VPCGatewayInternetGateway", cloudformation.Resource{
				Type: "AWS::EC2::InternetGateway",
			}))

			Expect(vpc.Resources).To(HaveKeyWithValue("VPCGatewayAttachment", cloudformation.Resource{
				Type: "AWS::EC2::VPCGatewayAttachment",
				Properties: cloudformation.VPCGatewayAttachment{
					VpcId:             cloudformation.Ref{"VPC"},
					InternetGatewayId: cloudformation.Ref{"VPCGatewayInternetGateway"},
				},
			}))

			Expect(vpc.Parameters).To(HaveLen(1))
			Expect(vpc.Parameters).To(HaveKeyWithValue("VPCCIDR", cloudformation.Parameter{
				Description: "CIDR block for the VPC.",
				Type:        "String",
				Default:     "10.0.0.0/16",
			}))
		})
	})

	Describe("BOSHSubnet", func() {
		It("returns a template with all fields for the BOSH subnet", func() {
			subnet := builder.BOSHSubnet()

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHSubnet", cloudformation.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: cloudformation.Subnet{
					VpcId:     cloudformation.Ref{"VPC"},
					CidrBlock: cloudformation.Ref{"BOSHSubnetCIDR"},
					Tags: []cloudformation.Tag{
						{
							Key:   "Name",
							Value: "BOSH",
						},
					},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHRouteTable", cloudformation.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: cloudformation.RouteTable{
					VpcId: cloudformation.Ref{"VPC"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHRoute", cloudformation.Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
				Properties: cloudformation.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            cloudformation.Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         cloudformation.Ref{"BOSHRouteTable"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("BOSHSubnetRouteTableAssociation", cloudformation.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: cloudformation.SubnetRouteTableAssociation{
					RouteTableId: cloudformation.Ref{"BOSHRouteTable"},
					SubnetId:     cloudformation.Ref{"BOSHSubnet"},
				},
			}))

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("BOSHSubnetCIDR", cloudformation.Parameter{
				Description: "CIDR block for the BOSH subnet.",
				Type:        "String",
				Default:     "10.0.0.0/24",
			}))
		})
	})

	Describe("InternalSubnet", func() {
		It("returns a template with all fields for the Internal subnet", func() {
			subnet := builder.InternalSubnet()

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("InternalSubnet", cloudformation.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: cloudformation.Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							"0",
							map[string]cloudformation.Ref{
								"Fn::GetAZs": cloudformation.Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: cloudformation.Ref{"InternalSubnetCIDR"},
					VpcId:     cloudformation.Ref{"VPC"},
					Tags: []cloudformation.Tag{
						{
							Key:   "Name",
							Value: "Internal",
						},
					},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("InternalRouteTable", cloudformation.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: cloudformation.RouteTable{
					VpcId: cloudformation.Ref{"VPC"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("InternalRoute", cloudformation.Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "NATInstance",
				Properties: cloudformation.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					RouteTableId:         cloudformation.Ref{"InternalRouteTable"},
					InstanceId:           cloudformation.Ref{"NATInstance"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("InternalSubnetRouteTableAssociation", cloudformation.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: cloudformation.SubnetRouteTableAssociation{
					RouteTableId: cloudformation.Ref{"InternalRouteTable"},
					SubnetId:     cloudformation.Ref{"InternalSubnet"},
				},
			}))

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("InternalSubnetCIDR", cloudformation.Parameter{
				Description: "CIDR block for the Internal subnet.",
				Type:        "String",
				Default:     "10.0.16.0/20",
			}))
		})
	})

	Describe("LoadBalancerSubnet", func() {
		It("returns a template with all fields for the load balancer subnet", func() {
			subnet := builder.LoadBalancerSubnet()

			Expect(subnet.Parameters).To(HaveLen(1))
			Expect(subnet.Parameters).To(HaveKeyWithValue("LoadBalancerSubnetCIDR", cloudformation.Parameter{
				Description: "CIDR block for the ELB subnet.",
				Type:        "String",
				Default:     "10.0.2.0/24",
			}))

			Expect(subnet.Resources).To(HaveLen(4))
			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerSubnet", cloudformation.Resource{
				Type: "AWS::EC2::Subnet",
				Properties: cloudformation.Subnet{
					AvailabilityZone: map[string]interface{}{
						"Fn::Select": []interface{}{
							"0",
							map[string]cloudformation.Ref{
								"Fn::GetAZs": cloudformation.Ref{"AWS::Region"},
							},
						},
					},
					CidrBlock: cloudformation.Ref{"LoadBalancerSubnetCIDR"},
					VpcId:     cloudformation.Ref{"VPC"},
					Tags: []cloudformation.Tag{
						{
							Key:   "Name",
							Value: "LoadBalancer",
						},
					},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerRouteTable", cloudformation.Resource{
				Type: "AWS::EC2::RouteTable",
				Properties: cloudformation.RouteTable{
					VpcId: cloudformation.Ref{"VPC"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerRoute", cloudformation.Resource{
				Type:      "AWS::EC2::Route",
				DependsOn: "VPCGatewayInternetGateway",
				Properties: cloudformation.Route{
					DestinationCidrBlock: "0.0.0.0/0",
					GatewayId:            cloudformation.Ref{"VPCGatewayInternetGateway"},
					RouteTableId:         cloudformation.Ref{"LoadBalancerRouteTable"},
				},
			}))

			Expect(subnet.Resources).To(HaveKeyWithValue("LoadBalancerSubnetRouteTableAssociation", cloudformation.Resource{
				Type: "AWS::EC2::SubnetRouteTableAssociation",
				Properties: cloudformation.SubnetRouteTableAssociation{
					RouteTableId: cloudformation.Ref{"LoadBalancerRouteTable"},
					SubnetId:     cloudformation.Ref{"LoadBalancerSubnet"},
				},
			}))
		})
	})

	Describe("Internal Security Group", func() {
		It("returns a template containing all the fields for internal security group", func() {
			security_group := builder.InternalSecurityGroup()

			Expect(security_group.Resources).To(HaveLen(5))
			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroup", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: cloudformation.SecurityGroup{
					VpcId:               cloudformation.Ref{"VPC"},
					GroupDescription:    "Internal",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []cloudformation.SecurityGroupIngress{
						{
							SourceSecurityGroupId: cloudformation.Ref{"WebSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
						{
							SourceSecurityGroupId: cloudformation.Ref{"WebSecurityGroup"},
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
			}))

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromBOSH", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: cloudformation.SecurityGroupIngress{
					GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: cloudformation.Ref{"BOSHSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromBOSH", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: cloudformation.SecurityGroupIngress{
					GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: cloudformation.Ref{"BOSHSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressTCPfromSelf", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: cloudformation.SecurityGroupIngress{
					GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
					IpProtocol:            "tcp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))

			Expect(security_group.Resources).To(HaveKeyWithValue("InternalSecurityGroupIngressUDPfromSelf", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroupIngress",
				Properties: cloudformation.SecurityGroupIngress{
					GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
					SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
					IpProtocol:            "udp",
					FromPort:              "0",
					ToPort:                "65535",
				},
			}))
		})
	})

	Describe("BOSH Security Group", func() {
		It("returns a template containing the bosh security group", func() {
			security_group := builder.BOSHSecurityGroup()

			Expect(security_group.Parameters).To(HaveLen(1))
			Expect(security_group.Parameters).To(HaveKeyWithValue("BOSHInboundCIDR", cloudformation.Parameter{
				Description: "CIDR to permit access to BOSH (e.g. 205.103.216.37/32 for your specific IP)",
				Type:        "String",
				Default:     "0.0.0.0/0",
			}))

			Expect(security_group.Resources).To(HaveLen(1))
			Expect(security_group.Resources).To(HaveKeyWithValue("BOSHSecurityGroup", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: cloudformation.SecurityGroup{
					VpcId:               cloudformation.Ref{"VPC"},
					GroupDescription:    "BOSH",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []cloudformation.SecurityGroupIngress{
						{
							CidrIp:     cloudformation.Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "22",
							ToPort:     "22",
						},

						{
							CidrIp:     cloudformation.Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "6868",
							ToPort:     "6868",
						},
						{
							CidrIp:     cloudformation.Ref{"BOSHInboundCIDR"},
							IpProtocol: "tcp",
							FromPort:   "25555",
							ToPort:     "25555",
						},
						{
							SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
						{
							SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
				},
			}))
		})
	})

	Describe("Web Security Group", func() {
		It("returns a template containing the web security group", func() {
			security_group := builder.WebSecurityGroup()

			Expect(security_group.Resources).To(HaveLen(1))
			Expect(security_group.Resources).To(HaveKeyWithValue("WebSecurityGroup", cloudformation.Resource{
				Type: "AWS::EC2::SecurityGroup",
				Properties: cloudformation.SecurityGroup{
					VpcId:               cloudformation.Ref{"VPC"},
					GroupDescription:    "Web",
					SecurityGroupEgress: []string{},
					SecurityGroupIngress: []cloudformation.SecurityGroupIngress{
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
			}))
		})
	})

	Describe("Web ELB Load Balancer", func() {
		It("returns a template containing the web elb load balancer", func() {
			web_elb_load_balancer := builder.WebELBLoadBalancer()

			Expect(web_elb_load_balancer.Resources).To(HaveLen(1))
			Expect(web_elb_load_balancer.Resources).To(HaveKeyWithValue("WebELBLoadBalancer", cloudformation.Resource{
				Type: "AWS::ElasticLoadBalancing::LoadBalancer",
				Properties: cloudformation.ElasticLoadBalancingLoadBalancer{
					Subnets:        []interface{}{cloudformation.Ref{"LoadBalancerSubnet"}},
					SecurityGroups: []interface{}{cloudformation.Ref{"WebSecurityGroup"}},

					HealthCheck: cloudformation.HealthCheck{
						HealthyThreshold:   "2",
						Interval:           "30",
						Target:             "tcp:8080",
						Timeout:            "5",
						UnhealthyThreshold: "10",
					},

					Listeners: []cloudformation.Listener{
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
			}))
		})
	})

	Describe("MicroEIP", func() {
		It("returns a template containing the micro elastic ip", func() {
			micro_eip := builder.MicroEIP()

			Expect(micro_eip.Resources).To(HaveLen(1))
			Expect(micro_eip.Resources).To(HaveKeyWithValue("MicroEIP", cloudformation.Resource{
				Type: "AWS::EC2::EIP",
				Properties: cloudformation.EIP{
					Domain: "vpc",
				},
			}))
		})
	})

	Describe("SSHKeyPairName", func() {
		It("returns a template containing the ssh keypair name", func() {
			ssh_keypair_name := builder.SSHKeyPairName("some-key-pair-name")

			Expect(ssh_keypair_name.Parameters).To(HaveLen(1))
			Expect(ssh_keypair_name.Parameters).To(HaveKeyWithValue("SSHKeyPairName", cloudformation.Parameter{
				Type:        "AWS::EC2::KeyPair::KeyName",
				Default:     "some-key-pair-name",
				Description: "SSH KeyPair to use for instances",
			}))
		})
	})

	Describe("template marshaling", func() {
		It("can be marshaled to JSON", func() {
			template := builder.Build("keypair-name")

			buf, err := ioutil.ReadFile("fixtures/cloudformation.json")
			Expect(err).NotTo(HaveOccurred())

			output, err := json.Marshal(template)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchJSON(string(buf)))
		})
	})
})
