package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MicroEIPTemplateBuilder", func() {
	var builder templates.MicroEIPTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewMicroEIPTemplateBuilder()
	})

	Describe("MicroEIP", func() {
		It("returns a template containing the micro elastic ip", func() {
			micro_eip := builder.MicroEIP()

			Expect(micro_eip.Resources).To(HaveLen(1))
			Expect(micro_eip.Resources).To(HaveKeyWithValue("MicroEIP", templates.Resource{
				Type: "AWS::EC2::EIP",
				Properties: templates.EIP{
					Domain: "vpc",
				},
			}))
		})
	})
})
