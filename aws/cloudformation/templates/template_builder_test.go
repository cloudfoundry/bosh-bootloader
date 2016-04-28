package templates_test

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateBuilder", func() {
	var (
		builder templates.TemplateBuilder
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		builder = templates.NewTemplateBuilder(logger)
	})

	Describe("Build", func() {
		Context("concourse elb template", func() {
			It("builds a cloudformation template", func() {
				template := builder.Build("keypair-name", 5, "concourse")
				Expect(template.AWSTemplateFormatVersion).To(Equal("2010-09-09"))
				Expect(template.Description).To(Equal("Infrastructure for a BOSH deployment with a Concourse ELB."))

				Expect(template.Parameters).To(HaveKey("SSHKeyPairName"))
				Expect(template.Resources).To(HaveKey("BOSHUser"))
				Expect(template.Resources).To(HaveKey("NATInstance"))
				Expect(template.Resources).To(HaveKey("VPC"))
				Expect(template.Resources).To(HaveKey("BOSHSubnet"))
				Expect(template.Resources).To(HaveKey("InternalSubnet1"))
				Expect(template.Resources).To(HaveKey("InternalSubnet2"))
				Expect(template.Resources).To(HaveKey("InternalSubnet3"))
				Expect(template.Resources).To(HaveKey("InternalSubnet4"))
				Expect(template.Resources).To(HaveKey("InternalSubnet5"))
				Expect(template.Resources).To(HaveKey("InternalSecurityGroup"))
				Expect(template.Resources).To(HaveKey("BOSHSecurityGroup"))
				Expect(template.Resources).To(HaveKey("BOSHEIP"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet1"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet2"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet3"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet4"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet5"))
				Expect(template.Resources).To(HaveKey("ConcourseSecurityGroup"))
				Expect(template.Resources).To(HaveKey("ConcourseInternalSecurityGroup"))
				Expect(template.Resources).To(HaveKey("ConcourseLoadBalancer"))
			})
		})

		Context("cf elb template", func() {
			It("builds a cloudformation template", func() {
				template := builder.Build("keypair-name", 5, "cf")
				Expect(template.AWSTemplateFormatVersion).To(Equal("2010-09-09"))
				Expect(template.Description).To(Equal("Infrastructure for a BOSH deployment with a CloudFoundry ELB."))

				Expect(template.Parameters).To(HaveKey("SSHKeyPairName"))
				Expect(template.Resources).To(HaveKey("BOSHUser"))
				Expect(template.Resources).To(HaveKey("NATInstance"))
				Expect(template.Resources).To(HaveKey("VPC"))
				Expect(template.Resources).To(HaveKey("BOSHSubnet"))
				Expect(template.Resources).To(HaveKey("InternalSubnet1"))
				Expect(template.Resources).To(HaveKey("InternalSubnet2"))
				Expect(template.Resources).To(HaveKey("InternalSubnet3"))
				Expect(template.Resources).To(HaveKey("InternalSubnet4"))
				Expect(template.Resources).To(HaveKey("InternalSubnet5"))
				Expect(template.Resources).To(HaveKey("InternalSecurityGroup"))
				Expect(template.Resources).To(HaveKey("BOSHSecurityGroup"))
				Expect(template.Resources).To(HaveKey("BOSHEIP"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet1"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet2"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet3"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet4"))
				Expect(template.Resources).To(HaveKey("LoadBalancerSubnet5"))
				Expect(template.Resources).To(HaveKey("CFRouterSecurityGroup"))
				Expect(template.Resources).To(HaveKey("CFRouterInternalSecurityGroup"))
				Expect(template.Resources).To(HaveKey("CFRouterLoadBalancer"))
				Expect(template.Resources).To(HaveKey("CFSSHProxySecurityGroup"))
				Expect(template.Resources).To(HaveKey("CFSSHProxyInternalSecurityGroup"))
				Expect(template.Resources).To(HaveKey("CFSSHProxyLoadBalancer"))
			})
		})

		Context("no elb template", func() {
			It("builds a cloudformation template", func() {
				template := builder.Build("keypair-name", 5, "")
				Expect(template.AWSTemplateFormatVersion).To(Equal("2010-09-09"))
				Expect(template.Description).To(Equal("Infrastructure for a BOSH deployment."))

				Expect(template.Parameters).To(HaveKey("SSHKeyPairName"))
				Expect(template.Resources).To(HaveKey("BOSHUser"))
				Expect(template.Resources).To(HaveKey("NATInstance"))
				Expect(template.Resources).To(HaveKey("VPC"))
				Expect(template.Resources).To(HaveKey("BOSHSubnet"))
				Expect(template.Resources).To(HaveKey("InternalSubnet1"))
				Expect(template.Resources).To(HaveKey("InternalSubnet2"))
				Expect(template.Resources).To(HaveKey("InternalSubnet3"))
				Expect(template.Resources).To(HaveKey("InternalSubnet4"))
				Expect(template.Resources).To(HaveKey("InternalSubnet5"))
				Expect(template.Resources).To(HaveKey("InternalSecurityGroup"))
				Expect(template.Resources).To(HaveKey("BOSHSecurityGroup"))
				Expect(template.Resources).To(HaveKey("BOSHEIP"))

				Expect(template.Resources).NotTo(HaveKey("LoadBalancerSubnet1"))
				Expect(template.Resources).NotTo(HaveKey("LoadBalancerSubnet2"))
				Expect(template.Resources).NotTo(HaveKey("LoadBalancerSubnet3"))
				Expect(template.Resources).NotTo(HaveKey("LoadBalancerSubnet4"))
				Expect(template.Resources).NotTo(HaveKey("LoadBalancerSubnet5"))
				Expect(template.Resources).NotTo(HaveKey("ConcourseSecurityGroup"))
				Expect(template.Resources).NotTo(HaveKey("ConcourseLoadBalancer"))
				Expect(template.Resources).NotTo(HaveKey("CFRouterSecurityGroup"))
				Expect(template.Resources).NotTo(HaveKey("CFRouterLoadBalancer"))
				Expect(template.Resources).NotTo(HaveKey("CFSSHProxySecurityGroup"))
				Expect(template.Resources).NotTo(HaveKey("CFSSHProxyLoadBalancer"))
			})
		})

		It("logs that the cloudformation template is being generated", func() {
			builder.Build("keypair-name", 0, "")

			Expect(logger.StepCall.Receives.Message).To(Equal("generating cloudformation template"))
		})
	})

	Describe("template marshaling", func() {
		DescribeTable("marshals template to JSON", func(lbType string, fixture string) {
			template := builder.Build("keypair-name", 4, lbType)

			buf, err := ioutil.ReadFile("fixtures/" + fixture)
			Expect(err).NotTo(HaveOccurred())

			output, err := json.Marshal(template)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchJSON(string(buf)))
		},
			Entry("without load balancer", "", "cloudformation_without_elb.json"),
			Entry("with cf load balancer", "cf", "cloudformation_with_cf_elb.json"),
			Entry("with concourse load balancer", "concourse", "cloudformation_with_concourse_elb.json"),
		)
	})
})
