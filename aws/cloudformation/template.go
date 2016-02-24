package cloudformation

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

type Template struct {
	AWSTemplateFormatVersion string                 `json:",omitempty"`
	Description              string                 `json:",omitempty"`
	Parameters               map[string]Parameter   `json:",omitempty"`
	Mappings                 map[string]interface{} `json:",omitempty"`
	Resources                map[string]Resource    `json:",omitempty"`
}

func (t Template) SetKeyPairName(name string) {
	t.Parameters["KeyName"] = Parameter{
		Type:        "AWS::EC2::KeyPair::KeyName",
		Default:     name,
		Description: "SSH Keypair to use for instances",
	}
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
	SecurityGroupEgress  []string
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

type VPC struct {
	CidrBlock Ref   `json:",omitempty"`
	Tags      []Tag `json:",omitempty"`
}

type VPCGatewayAttachment struct {
	VpcId             Ref `json:",omitempty"`
	InternetGatewayId Ref `json:",omitempty"`
}

type ElasticLoadBalancingLoadBalancer struct {
	Subnets        []interface{} `json:",omitempty"`
	SecurityGroups []interface{} `json:",omitempty"`
	HealthCheck    HealthCheck   `json:",omitempty"`
	Listeners      []Listener    `json:',omitempty"`
}

type Listener struct {
	Protocol         string                 `json:",omitempty"`
	LoadBalancerPort string                 `json:",omitempty"`
	InstanceProtocol string                 `json:",omitempty"`
	InstancePort     string                 `json:",omitempty"`
	SSLCertificateId map[string]interface{} `json:",omitempty"`
}

type HealthCheck struct {
	HealthyThreshold   string `json",omitempty"`
	Interval           string `json",omitempty"`
	Target             string `json",omitempty"`
	Timeout            string `json",omitempty"`
	UnhealthyThreshold string `json",omitempty"`
}
