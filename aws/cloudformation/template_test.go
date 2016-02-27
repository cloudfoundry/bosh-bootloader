package cloudformation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
)

var _ = Describe("Template", func() {
	Context("SetKeyPairName", func() {
		It("sets the keypair parameter to the given name", func() {
			builder := cloudformation.TemplateBuilder{}
			template := builder.Build()

			Expect(template.Parameters["KeyName"].Default).NotTo(Equal("some-keypair-name"))

			template.SetKeyPairName("some-keypair-name")

			Expect(template.Parameters["KeyName"]).To(Equal(
				cloudformation.Parameter{
					Type:        "AWS::EC2::KeyPair::KeyName",
					Default:     "some-keypair-name",
					Description: "SSH KeyPair to use for instances",
				},
			))
		})
	})
})
