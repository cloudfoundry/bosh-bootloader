package renderers_test

import (
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe(renderers.ShellTypePowershell, func() {
	var (
		renderer renderers.Renderer
	)

	BeforeEach(func() {
		renderer = renderers.NewPowershell()
	})

	Describe("RenderEnvironmentVariable", func() {
		Context("WhenSingleLine", func() {
			It("prints env statement properly", func() {
				key := "KEY"
				value := "value"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("$env:KEY=\"value\""))
			})
		})
		Context("WhenMultiLine", func() {
			It("prints env statement with enclosing quotes", func() {
				key := "KEY"
				value := "1\r\n2\r\n3\r\n4\r\n"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("$env:KEY='\r\n1\r\n2\r\n3\r\n4\r\n'"))
			})
			It("appends newline if not present", func() {
				key := "KEY"
				value := "1\r\n2\r\n3\r\n4"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("$env:KEY='\r\n1\r\n2\r\n3\r\n4\r\n'"))
			})
		})
	})

	Describe("Type", func() {
		It("is powershell", func() {
			t := renderer.Type()
			Expect(t).To(Equal(renderers.ShellTypePowershell))
		})
	})
})
