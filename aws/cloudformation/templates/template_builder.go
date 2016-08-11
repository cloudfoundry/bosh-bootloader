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

func (t TemplateBuilder) Build(keyPairName string, numberOfAvailabilityZones int,
	lbType, lbCertificateARN, envID string) Template {
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
		internalSubnetsTemplateBuilder.InternalSubnets(numberOfAvailabilityZones, envID),
		sshKeyPairTemplateBuilder.SSHKeyPairName(keyPairName),
		boshIAMTemplateBuilder.BOSHIAMUser(),
		natTemplateBuilder.NAT(envID),
		vpcTemplateBuilder.VPC(envID),
		boshSubnetTemplateBuilder.BOSHSubnet(envID),
		securityGroupTemplateBuilder.InternalSecurityGroup(envID),
		securityGroupTemplateBuilder.BOSHSecurityGroup(envID),
		boshEIPTemplateBuilder.BOSHEIP(),
	)

	if lbType == "concourse" {
		template.Description = "Infrastructure for a BOSH deployment with a Concourse ELB."

		lbTemplate := loadBalancerTemplateBuilder.ConcourseLoadBalancer(numberOfAvailabilityZones, lbCertificateARN, envID)
		template.Merge(
			loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets(numberOfAvailabilityZones, envID),
			lbTemplate,
			securityGroupTemplateBuilder.LBSecurityGroup("ConcourseSecurityGroup", "Concourse", "ConcourseLoadBalancer", lbTemplate, envID),
			securityGroupTemplateBuilder.LBInternalSecurityGroup("ConcourseInternalSecurityGroup", "ConcourseSecurityGroup", "ConcourseInternal", "ConcourseLoadBalancer", lbTemplate, envID),
		)
	}

	if lbType == "cf" {
		template.Description = "Infrastructure for a BOSH deployment with a CloudFoundry ELB."
		routerLBTemplate := loadBalancerTemplateBuilder.CFRouterLoadBalancer(numberOfAvailabilityZones, lbCertificateARN, envID)
		sshLBTemplate := loadBalancerTemplateBuilder.CFSSHProxyLoadBalancer(numberOfAvailabilityZones, envID)
		template.Merge(
			loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets(numberOfAvailabilityZones, envID),

			routerLBTemplate,
			securityGroupTemplateBuilder.LBSecurityGroup("CFRouterSecurityGroup", "Router", "CFRouterLoadBalancer", routerLBTemplate, envID),
			securityGroupTemplateBuilder.LBInternalSecurityGroup("CFRouterInternalSecurityGroup", "CFRouterSecurityGroup", "CFRouterInternal", "CFRouterLoadBalancer", routerLBTemplate, envID),

			sshLBTemplate,
			securityGroupTemplateBuilder.LBSecurityGroup("CFSSHProxySecurityGroup", "CFSSHProxy", "CFSSHProxyLoadBalancer", sshLBTemplate, envID),
			securityGroupTemplateBuilder.LBInternalSecurityGroup("CFSSHProxyInternalSecurityGroup", "CFSSHProxySecurityGroup", "CFSSHProxyInternal", "CFSSHProxyLoadBalancer", sshLBTemplate, envID),
		)
	}

	return template
}
