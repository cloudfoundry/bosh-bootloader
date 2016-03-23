package templates

import "fmt"

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

func (t TemplateBuilder) Build(keyPairName string, availabilityZones []string) Template {
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
		t.subnetTemplates(availabilityZones),
		sshKeyPairTemplateBuilder.SSHKeyPairName(keyPairName),
		boshIAMTemplateBuilder.BOSHIAMUser(),
		natTemplateBuilder.NAT(),
		vpcTemplateBuilder.VPC(),
		subnetTemplateBuilder.BOSHSubnet(),
		subnetTemplateBuilder.LoadBalancerSubnet(),
		securityGroupTemplateBuilder.InternalSecurityGroup(),
		securityGroupTemplateBuilder.BOSHSecurityGroup(),
		securityGroupTemplateBuilder.WebSecurityGroup(),
		webELBTemplateBuilder.WebELBLoadBalancer(),
		boshEIPTemplateBuilder.BOSHEIP(),
	)
}

func (t TemplateBuilder) subnetTemplates(availabilityZones []string) Template {
	template := Template{}

	subnetTemplateBuilder := NewSubnetTemplateBuilder()
	for index, _ := range availabilityZones {
		template = template.Merge(subnetTemplateBuilder.InternalSubnet(
			index,
			fmt.Sprintf("%d", index+1),
			fmt.Sprintf("10.0.%d.0/20", 16*(index+1)),
		))
	}

	return template
}
