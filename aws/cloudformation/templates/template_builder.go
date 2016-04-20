package templates

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

func (t TemplateBuilder) Build(keyPairName string, numberOfAvailabilityZones int, lbType string) Template {
	t.logger.Step("generating cloudformation template")

	boshIAMTemplateBuilder := NewBOSHIAMTemplateBuilder()
	natTemplateBuilder := NewNATTemplateBuilder()
	vpcTemplateBuilder := NewVPCTemplateBuilder()
	internalSubnetsTemplateBuilder := NewInternalSubnetsTemplateBuilder()
	boshSubnetTemplateBuilder := NewBOSHSubnetTemplateBuilder()
	boshEIPTemplateBuilder := NewBOSHEIPTemplateBuilder()
	securityGroupTemplateBuilder := NewSecurityGroupTemplateBuilder()
	sshKeyPairTemplateBuilder := NewSSHKeyPairTemplateBuilder()
	loadBalancerSubnetsTemplateBuilder := NewLoadBalancerSubnetsTemplateBuilder()
	concourseELBTemplateBuilder := NewConcourseELBTemplateBuilder()
	cfLoadBalancerTemplateBuilder := NewCFLoadBalancerTemplateBuilder()

	template := Template{
		AWSTemplateFormatVersion: "2010-09-09",
		Description:              "Infrastructure for a BOSH deployment.",
	}.Merge(
		internalSubnetsTemplateBuilder.InternalSubnets(numberOfAvailabilityZones),
		sshKeyPairTemplateBuilder.SSHKeyPairName(keyPairName),
		boshIAMTemplateBuilder.BOSHIAMUser(),
		natTemplateBuilder.NAT(),
		vpcTemplateBuilder.VPC(),
		boshSubnetTemplateBuilder.BOSHSubnet(),
		securityGroupTemplateBuilder.InternalSecurityGroup(),
		securityGroupTemplateBuilder.BOSHSecurityGroup(),
		boshEIPTemplateBuilder.BOSHEIP(),
	)

	if lbType == "concourse" {
		template.Description = "Infrastructure for a BOSH deployment with a Concourse ELB."
		template.Merge(
			loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets(numberOfAvailabilityZones),
			securityGroupTemplateBuilder.ConcourseSecurityGroup(),
			concourseELBTemplateBuilder.ConcourseLoadBalancer(numberOfAvailabilityZones),
		)
	}

	if lbType == "cf" {
		template.Description = "Infrastructure for a BOSH deployment with a CloudFoundry ELB."
		template.Merge(
			loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets(numberOfAvailabilityZones),
			securityGroupTemplateBuilder.RouterSecurityGroup(),
			cfLoadBalancerTemplateBuilder.CFLoadBalancer(numberOfAvailabilityZones),
		)
	}

	return template
}
