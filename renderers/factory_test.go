package renderers_test

import (
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {
	Describe("Create", func() {
		var (
			factory renderers.Factory
			vars    map[string]string
		)
		BeforeEach(func() {
			vars = make(map[string]string)
		})
		Context("WhenPSModulePathSet", func() {
			BeforeEach(func() {
				envGetter := &fakes.EnvGetter{Values: vars}
				factory = renderers.NewFactory(envGetter)
			})
			It("creates powershell renderer", func() {
				vars["PSModulePath"] = "anything"
				shellType := ""
				renderer, err := factory.Create(shellType)
				Expect(err).To(BeNil())
				Expect(renderer.Type()).To(Equal(renderers.ShellTypePowershell))
			})
		})
		Context("WhenPSModulePathUnset", func() {
			BeforeEach(func() {
				envGetter := &fakes.EnvGetter{Values: vars}
				factory = renderers.NewFactory(envGetter)
			})
			It("creates posix renderer", func() {
				shellType := ""
				renderer, err := factory.Create(shellType)
				Expect(err).To(BeNil())
				Expect(renderer.Type()).To(Equal(renderers.ShellTypePosix))
			})
		})
		Context("When passed 'yaml'", func() {
			BeforeEach(func() {
				envGetter := &fakes.EnvGetter{Values: vars}
				factory = renderers.NewFactory(envGetter)
			})
			It("creates yaml renderer", func() {
				shellType := "yaml"
				renderer, err := factory.Create(shellType)
				Expect(err).To(BeNil())
				Expect(renderer.Type()).To(Equal(renderers.ShellTypeYaml))
			})
		})
	})
})
