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

func (t TemplateBuilder) Build(keyPairName string) Template {
	t.logger.Step("generating cloudformation template")

	boshIAMTemplateBuilder := NewBOSHIAMTemplateBuilder()
	natTemplateBuilder := NewNATTemplateBuilder()
	vpcTemplateBuilder := NewVPCTemplateBuilder()
	subnetTemplateBuilder := NewSubnetTemplateBuilder()
	securityGroupTemplateBuilder := NewSecurityGroupTemplateBuilder()
	webELBTemplateBuilder := NewWebELBTemplateBuilder()
	boshEIPTemplateBuilder := NewBOSHEIPTemplateBuilder()
	sshKeyPairTemplateBuilder := NewSSHKeyPairTemplateBuilder()

	return Template{
		AWSTemplateFormatVersion: "2010-09-09",
		Description:              "Infrastructure for a BOSH deployment with an ELB.",
	}.Merge(
		sshKeyPairTemplateBuilder.SSHKeyPairName(keyPairName),
		boshIAMTemplateBuilder.BOSHIAMUser(),
		natTemplateBuilder.NAT(),
		vpcTemplateBuilder.VPC(),
		subnetTemplateBuilder.BOSHSubnet(),
		subnetTemplateBuilder.InternalSubnet(0, "1", "10.0.16.0/20"),
		subnetTemplateBuilder.InternalSubnet(1, "2", "10.0.32.0/20"),
		subnetTemplateBuilder.InternalSubnet(2, "3", "10.0.48.0/20"),
		subnetTemplateBuilder.LoadBalancerSubnet(),
		securityGroupTemplateBuilder.InternalSecurityGroup(),
		securityGroupTemplateBuilder.BOSHSecurityGroup(),
		securityGroupTemplateBuilder.WebSecurityGroup(),
		webELBTemplateBuilder.WebELBLoadBalancer(),
		boshEIPTemplateBuilder.BOSHEIP(),
	)
}
