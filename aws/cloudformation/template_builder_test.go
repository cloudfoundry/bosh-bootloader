package cloudformation_test

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateBuilder", func() {
	var builder cloudformation.TemplateBuilder

	BeforeEach(func() {
		builder = cloudformation.NewTemplateBuilder()
	})

	Describe("Build", func() {
		It("builds a cloudformation template", func() {
			template := builder.Build("keypair-name")
			Expect(template).To(Equal(cloudformation.Template{
				AWSTemplateFormatVersion: "2010-09-09",
				Description:              "Infrastructure for a MicroBOSH deployment with an ELB.",
				Parameters: map[string]cloudformation.Parameter{
					"KeyName": {
						Type:        "AWS::EC2::KeyPair::KeyName",
						Default:     "keypair-name",
						Description: "SSH KeyPair to use for instances",
					},
					"BOSHInboundCIDR": {
						Description: "CIDR to permit access to BOSH (e.g. 205.103.216.37/32 for your specific IP)",
						Type:        "String",
						Default:     "0.0.0.0/0",
					},

					"VPCCIDR": {
						Description: "CIDR block for the VPC.",
						Type:        "String",
						Default:     "10.0.0.0/16",
					},
					"BOSHSubnetCIDR": {
						Description: "CIDR block for the BOSH subnet.",
						Type:        "String",
						Default:     "10.0.0.0/24",
					},
					"LoadBalancerSubnetCIDR": {
						Description: "CIDR block for the BOSH subnet.",
						Type:        "String",
						Default:     "10.0.2.0/24",
					},
					"InternalSubnetCIDR": {
						Description: "CIDR block for the Internal subnet.",
						Type:        "String",
						Default:     "10.0.16.0/20",
					},
				},
				Mappings: map[string]interface{}{
					"AWSNATAMI": map[string]cloudformation.AMI{
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
				Resources: map[string]cloudformation.Resource{
					"VPC": {
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
					},
					"VPCGatewayInternetGateway": {
						Type: "AWS::EC2::InternetGateway",
					},
					"VPCGatewayAttachment": {
						Type: "AWS::EC2::VPCGatewayAttachment",
						Properties: cloudformation.VPCGatewayAttachment{
							VpcId:             cloudformation.Ref{"VPC"},
							InternetGatewayId: cloudformation.Ref{"VPCGatewayInternetGateway"},
						},
					},
					"BOSHSecurityGroup": {
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
					},
					"InternalSecurityGroup": {
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
					},
					"InternalSecurityGroupIngressTCPfromBOSH": {
						Type: "AWS::EC2::SecurityGroupIngress",
						Properties: cloudformation.SecurityGroupIngress{
							GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
							SourceSecurityGroupId: cloudformation.Ref{"BOSHSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
					"InternalSecurityGroupIngressUDPfromBOSH": {
						Type: "AWS::EC2::SecurityGroupIngress",
						Properties: cloudformation.SecurityGroupIngress{
							GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
							SourceSecurityGroupId: cloudformation.Ref{"BOSHSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
					"InternalSecurityGroupIngressTCPfromSelf": {
						Type: "AWS::EC2::SecurityGroupIngress",
						Properties: cloudformation.SecurityGroupIngress{
							GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
							SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
							IpProtocol:            "tcp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
					"InternalSecurityGroupIngressUDPfromSelf": {
						Type: "AWS::EC2::SecurityGroupIngress",
						Properties: cloudformation.SecurityGroupIngress{
							GroupId:               cloudformation.Ref{"InternalSecurityGroup"},
							SourceSecurityGroupId: cloudformation.Ref{"InternalSecurityGroup"},
							IpProtocol:            "udp",
							FromPort:              "0",
							ToPort:                "65535",
						},
					},
					"WebSecurityGroup": {
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
					},
					"NATSecurityGroup": {
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
					},
					"BOSHSubnet": {
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
					},
					"BOSHRouteTable": {
						Type: "AWS::EC2::RouteTable",
						Properties: cloudformation.RouteTable{
							VpcId: cloudformation.Ref{"VPC"},
						},
					},
					"BOSHRoute": {
						Type:      "AWS::EC2::Route",
						DependsOn: "VPCGatewayInternetGateway",
						Properties: cloudformation.Route{
							DestinationCidrBlock: "0.0.0.0/0",
							GatewayId:            cloudformation.Ref{"VPCGatewayInternetGateway"},
							RouteTableId:         cloudformation.Ref{"BOSHRouteTable"},
						},
					},
					"BOSHSubnetRouteTableAssociation": {
						Type: "AWS::EC2::SubnetRouteTableAssociation",
						Properties: cloudformation.SubnetRouteTableAssociation{
							RouteTableId: cloudformation.Ref{"BOSHRouteTable"},
							SubnetId:     cloudformation.Ref{"BOSHSubnet"},
						},
					},
					"NATInstance": {
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
							KeyName: cloudformation.Ref{"KeyName"},
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
					},
					"NATEIP": {
						Type: "AWS::EC2::EIP",
						Properties: cloudformation.EIP{
							Domain:     "vpc",
							InstanceId: cloudformation.Ref{"NATInstance"},
						},
					},
					"InternalSubnet": {
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
					},
					"InternalRouteTable": {
						Type: "AWS::EC2::RouteTable",
						Properties: cloudformation.RouteTable{
							VpcId: cloudformation.Ref{"VPC"},
						},
					},
					"InternalRoute": {
						Type:      "AWS::EC2::Route",
						DependsOn: "NATInstance",
						Properties: cloudformation.Route{
							DestinationCidrBlock: "0.0.0.0/0",
							RouteTableId:         cloudformation.Ref{"InternalRouteTable"},
							InstanceId:           cloudformation.Ref{"NATInstance"},
						},
					},
					"InternalSubnetRouteTableAssociation": {
						Type: "AWS::EC2::SubnetRouteTableAssociation",
						Properties: cloudformation.SubnetRouteTableAssociation{
							RouteTableId: cloudformation.Ref{"InternalRouteTable"},
							SubnetId:     cloudformation.Ref{"InternalSubnet"},
						},
					},
					"LoadBalancerSubnet": {
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
					},
					"LoadBalancerRouteTable": {
						Type: "AWS::EC2::RouteTable",
						Properties: cloudformation.RouteTable{
							VpcId: cloudformation.Ref{"VPC"},
						},
					},
					"LoadBalancerRoute": {
						Type:      "AWS::EC2::Route",
						DependsOn: "VPCGatewayInternetGateway",
						Properties: cloudformation.Route{
							DestinationCidrBlock: "0.0.0.0/0",
							GatewayId:            cloudformation.Ref{"VPCGatewayInternetGateway"},
							RouteTableId:         cloudformation.Ref{"LoadBalancerRouteTable"},
						},
					},
					"LoadBalancerSubnetRouteTableAssociation": {
						Type: "AWS::EC2::SubnetRouteTableAssociation",
						Properties: cloudformation.SubnetRouteTableAssociation{
							RouteTableId: cloudformation.Ref{"LoadBalancerRouteTable"},
							SubnetId:     cloudformation.Ref{"LoadBalancerSubnet"},
						},
					},
					"WebELBLoadBalancer": {
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
					},
					"MicroEIP": {
						Type: "AWS::EC2::EIP",
						Properties: cloudformation.EIP{
							Domain: "vpc",
						},
					},
				},
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
