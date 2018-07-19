package renderers_test

import (
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bash", func() {
	var (
		renderer renderers.Renderer
	)

	BeforeEach(func() {
		renderer = renderers.NewBash()
	})

	Describe("RenderEnvironmentVariable", func() {
		Context("WhenSingleLine", func() {
			It("prints env statement properly", func() {
				key := "KEY"
				value := "value"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("export KEY=value"))
			})
		})
		Context("WhenMultiLine", func() {
			It("prints env statement with enclosing quotes", func() {
				key := "KEY"
				value := "1\n2\n3\n4\n"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("export KEY='1\n2\n3\n4\n'"))
			})
			It("appends newline if not present", func() {
				key := "KEY"
				value := "1\n2\n3\n4"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("export KEY='1\n2\n3\n4\n'"))
			})
		})
	})

	Describe("RenderEnvironment", func() {
		Context("WhenMultiple", func() {
			It("prints each on a line", func() {
				result := renderer.RenderEnvironment(
					map[string]string{
						"KEY":   "value",
						"OTHER": "other",
					})
				Expect(result).To(Equal("export KEY=value\nexport OTHER=other\n"))
			})
		})
	})

	Describe("Shell", func() {
		It("is bash", func() {
			shell := renderer.Shell()
			Expect(shell).To(Equal("bash"))
		})
	})
})
