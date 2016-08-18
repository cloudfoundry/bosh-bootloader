package templates

import "encoding/json"

type AMI struct {
	AMI string `json:",omitempty"`
}

type Ref struct {
	Ref string `json:",omitempty"`
}

type Tag struct {
	Key   string `json:",omitempty"`
	Value string `json:",omitempty"`
}

type FnGetAtt struct {
	FnGetAtt []string `json:"Fn::GetAtt"`
}

type FnJoin struct {
	Delimeter string
	Values    []interface{}
}

func (j FnJoin) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]interface{}{
		"Fn::Join": {j.Delimeter, j.Values},
	})
}

type Template struct {
	AWSTemplateFormatVersion string                 `json:",omitempty"`
	Description              string                 `json:",omitempty"`
	Parameters               map[string]Parameter   `json:",omitempty"`
	Mappings                 map[string]interface{} `json:",omitempty"`
	Resources                map[string]Resource    `json:",omitempty"`
	Outputs                  map[string]Output      `json:",omitempty"`
}

func (t Template) Merge(templates ...Template) Template {
	if t.Parameters == nil {
		t.Parameters = map[string]Parameter{}
	}

	if t.Mappings == nil {
		t.Mappings = map[string]interface{}{}
	}

	if t.Resources == nil {
		t.Resources = map[string]Resource{}
	}

	if t.Outputs == nil {
		t.Outputs = map[string]Output{}
	}

	for _, template := range templates {
		for name, parameter := range template.Parameters {
			t.Parameters[name] = parameter
		}

		for name, mapping := range template.Mappings {
			t.Mappings[name] = mapping
		}

		for name, resource := range template.Resources {
			t.Resources[name] = resource
		}

		for name, output := range template.Outputs {
			t.Outputs[name] = output
		}
	}

	return t
}

type Output struct {
	Value interface{}
}

type Parameter struct {
	Type        string
	Default     string
	Description string `json:",omitempty"`
}

type Resource struct {
	Type           string
	Properties     interface{} `json:",omitempty"`
	DependsOn      interface{} `json:",omitempty"`
	CreationPolicy interface{} `json:",omitempty"`
	UpdatePolicy   interface{} `json:",omitempty"`
	DeletionPolicy interface{} `json:",omitempty"`
}

type SecurityGroup struct {
	VpcId                interface{}            `json:",omitempty"`
	GroupDescription     string                 `json:",omitempty"`
	SecurityGroupIngress []SecurityGroupIngress `json:",omitempty"`
	SecurityGroupEgress  []SecurityGroupEgress
}

type SecurityGroupEgress struct {
	SourceSecurityGroupId interface{} `json:",omitempty"`
	IpProtocol            string      `json:",omitempty"`
	FromPort              string      `json:",omitempty"`
	ToPort                string      `json:",omitempty"`
}

type SecurityGroupIngress struct {
	GroupId               interface{} `json:",omitempty"`
	SourceSecurityGroupId interface{} `json:",omitempty"`
	CidrIp                interface{} `json:",omitempty"`
	IpProtocol            string      `json:",omitempty"`
	FromPort              string      `json:",omitempty"`
	ToPort                string      `json:",omitempty"`
}

type SubnetRouteTableAssociation struct {
	RouteTableId interface{} `json:",omitempty"`
	SubnetId     interface{} `json:",omitempty"`
}

type Route struct {
	DestinationCidrBlock string      `json:",omitempty"`
	GatewayId            interface{} `json:",omitempty"`
	RouteTableId         interface{} `json:",omitempty"`
	InstanceId           interface{} `json:",omitempty"`
}

type Instance struct {
	InstanceType     string                 `json:",omitempty"`
	SubnetId         interface{}            `json:",omitempty"`
	ImageId          map[string]interface{} `json:",omitempty"`
	KeyName          interface{}            `json:",omitempty"`
	SecurityGroupIds []interface{}          `json:",omitempty"`
	Tags             []Tag                  `json:",omitempty"`
	SourceDestCheck  bool
}

type EIP struct {
	Domain     string      `json:",omitempty"`
	InstanceId interface{} `json:",omitempty"`
}

type Subnet struct {
	AvailabilityZone map[string]interface{} `json:",omitempty"`
	CidrBlock        interface{}            `json:",omitempty"`
	VpcId            interface{}            `json:",omitempty"`
	Tags             []Tag                  `json:",omitempty"`
}

type RouteTable struct {
	VpcId interface{} `json:",omitempty"`
}

type IAMUser struct {
	Policies []IAMPolicy
	UserName string `json:",omitempty"`
}

type IAMPolicy struct {
	PolicyName     string
	PolicyDocument IAMPolicyDocument
}

type IAMPolicyDocument struct {
	Version   string
	Statement []IAMStatement
}

type IAMStatement struct {
	Action   []string
	Effect   string
	Resource string
}

type IAMAccessKey struct {
	UserName Ref
}

type VPC struct {
	CidrBlock Ref   `json:"CidrBlock,omitempty"`
	Tags      []Tag `json:"Tags,omitempty"`
}

type VPCGatewayAttachment struct {
	VpcId             Ref `json:"VpcId,omitempty"`
	InternetGatewayId Ref `json:"InternetGatewayId,omitempty"`
}

type ElasticLoadBalancingLoadBalancer struct {
	Subnets        []interface{} `json:"Subnets,omitempty"`
	SecurityGroups []interface{} `json:"SecurityGroups,omitempty"`
	HealthCheck    HealthCheck   `json:"HealthCheck,omitempty"`
	Listeners      []Listener    `json:"Listeners,omitempty"`
	CrossZone      bool          `json:"CrossZone,omitempty"`
}

type Listener struct {
	Protocol         string `json:"Protocol,omitempty"`
	LoadBalancerPort string `json:"LoadBalancerPort,omitempty"`
	InstanceProtocol string `json:"InstanceProtocol,omitempty"`
	InstancePort     string `json:"InstancePort,omitempty"`
	SSLCertificateID string `json:"SSLCertificateId,omitempty"`
}

type HealthCheck struct {
	HealthyThreshold   string `json:"HealthyThreshold,omitempty"`
	Interval           string `json:"Interval,omitempty"`
	Target             string `json:"Target,omitempty"`
	Timeout            string `json:"Timeout,omitempty"`
	UnhealthyThreshold string `json:"UnhealthyThreshold,omitempty"`
}
