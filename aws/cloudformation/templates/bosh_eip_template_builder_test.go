package templates_test

import (
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSHEIPTemplateBuilder", func() {
	var builder templates.BOSHEIPTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewBOSHEIPTemplateBuilder()
	})

	Describe("BOSHEIP", func() {
		It("returns a template containing the bosh elastic ip", func() {
			eip := builder.BOSHEIP()

			Expect(eip.Resources).To(HaveLen(1))
			Expect(eip.Resources).To(HaveKeyWithValue("BOSHEIP", templates.Resource{
				DependsOn: "VPCGatewayAttachment",
				Type:      "AWS::EC2::EIP",
				Properties: templates.EIP{
					Domain: "vpc",
				},
			}))

			Expect(eip.Outputs).To(HaveLen(2))
			Expect(eip.Outputs).To(HaveKeyWithValue("BOSHEIP", templates.Output{
				Value: templates.Ref{"BOSHEIP"},
			}))
			Expect(eip.Outputs).To(HaveKeyWithValue("BOSHURL", templates.Output{
				Value: templates.FnJoin{
					Delimeter: "",
					Values:    []interface{}{"https://", templates.Ref{"BOSHEIP"}, ":25555"},
				},
			}))
		})
	})
})
