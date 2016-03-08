package templates_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSHEIPTemplateBuilder", func() {
	var builder templates.BOSHEIPTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewBOSHEIPTemplateBuilder()
	})

	Describe("BOSHEIP", func() {
		It("returns a template containing the micro elastic ip", func() {
			micro_eip := builder.BOSHEIP()

			Expect(micro_eip.Resources).To(HaveLen(1))
			Expect(micro_eip.Resources).To(HaveKeyWithValue("BOSHEIP", templates.Resource{
				Type: "AWS::EC2::EIP",
				Properties: templates.EIP{
					Domain: "vpc",
				},
			}))
		})
	})
})
