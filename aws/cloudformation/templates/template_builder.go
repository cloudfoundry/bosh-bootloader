package templates

type logger interface {
	Step(message string, a ...interface{})
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

func (t TemplateBuilder) Build(keyPairName string, numberOfAvailabilityZones int, lbType, lbCertificateARN string, iamUserName string, envID string) Template {
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
	loadBalancerTemplateBuilder := NewLoadBalancerTemplateBuilder()

	template := Template{
		AWSTemplateFormatVersion: "2010-09-09",
		Description:              "Infrastructure for a BOSH deployment.",
	}.Merge(
		internalSubnetsTemplateBuilder.InternalSubnets(numberOfAvailabilityZones),
		sshKeyPairTemplateBuilder.SSHKeyPairName(keyPairName),
		boshIAMTemplateBuilder.BOSHIAMUser(iamUserName),
		natTemplateBuilder.NAT(),
		vpcTemplateBuilder.VPC(envID),
		boshSubnetTemplateBuilder.BOSHSubnet(),
		securityGroupTemplateBuilder.InternalSecurityGroup(),
		securityGroupTemplateBuilder.BOSHSecurityGroup(),
		boshEIPTemplateBuilder.BOSHEIP(),
	)

	if lbType == "concourse" {
		template.Description = "Infrastructure for a BOSH deployment with a Concourse ELB."

		lbTemplate := loadBalancerTemplateBuilder.ConcourseLoadBalancer(numberOfAvailabilityZones, lbCertificateARN)
		template.Merge(
			loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets(numberOfAvailabilityZones),
			lbTemplate,
			securityGroupTemplateBuilder.LBSecurityGroup("ConcourseSecurityGroup", "Concourse", "ConcourseLoadBalancer", lbTemplate),
			securityGroupTemplateBuilder.LBInternalSecurityGroup("ConcourseInternalSecurityGroup", "ConcourseSecurityGroup", "ConcourseInternal", "ConcourseLoadBalancer", lbTemplate),
		)
	}

	if lbType == "cf" {
		template.Description = "Infrastructure for a BOSH deployment with a CloudFoundry ELB."
		routerLBTemplate := loadBalancerTemplateBuilder.CFRouterLoadBalancer(numberOfAvailabilityZones, lbCertificateARN)
		sshLBTemplate := loadBalancerTemplateBuilder.CFSSHProxyLoadBalancer(numberOfAvailabilityZones)
		template.Merge(
			loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets(numberOfAvailabilityZones),

			routerLBTemplate,
			securityGroupTemplateBuilder.LBSecurityGroup("CFRouterSecurityGroup", "Router", "CFRouterLoadBalancer", routerLBTemplate),
			securityGroupTemplateBuilder.LBInternalSecurityGroup("CFRouterInternalSecurityGroup", "CFRouterSecurityGroup", "CFRouterInternal", "CFRouterLoadBalancer", routerLBTemplate),

			sshLBTemplate,
			securityGroupTemplateBuilder.LBSecurityGroup("CFSSHProxySecurityGroup", "CFSSHProxy", "CFSSHProxyLoadBalancer", sshLBTemplate),
			securityGroupTemplateBuilder.LBInternalSecurityGroup("CFSSHProxyInternalSecurityGroup", "CFSSHProxySecurityGroup", "CFSSHProxyInternal", "CFSSHProxyLoadBalancer", sshLBTemplate),
		)
	}

	return template
}
