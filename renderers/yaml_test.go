package renderers_test

import (
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe(renderers.ShellTypeYaml, func() {
	var (
		renderer renderers.Renderer
	)

	BeforeEach(func() {
		renderer = renderers.NewYaml()
	})

	Describe("RenderEnvironmentVariable", func() {
		Context("WhenSingleLine", func() {
			It("prints env statement properly", func() {
				key := "KEY"
				value := "value"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal(`key: "value"`))
			})
		})
		Context("WhenMultiLine", func() {
			It("prints env statement with enclosing quotes", func() {
				key := "KEY"
				value := "1\n2\n3\n4\n"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal(`key: "1\n2\n3\n4\n"`))
			})
			It("appends newline if not present", func() {
				key := "KEY"
				value := "1\n2\n3\n4"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal(`key: "1\n2\n3\n4\n"`))
			})
		})
	})

	Describe("Type", func() {
		It("is yaml", func() {
			shellType := renderer.Type()
			Expect(shellType).To(Equal(renderers.ShellTypeYaml))
		})
	})
})
