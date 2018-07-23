package renderers_test

import (
	"os"

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
		Context("WhenWindowsPlatform", func() {
			BeforeEach(func() {
				platform := "windows"
				envGetter := &fakes.EnvGetter{Values: vars}
				factory = renderers.NewFactory(platform, envGetter)
			})
			Context("WhenDefaultShellType", func() {
				It("creates powershell renderer", func() {
					shellType := ""
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(renderers.ShellTypePowershell))
				})
			})
			Context("WhenShellTypeIsPowershell", func() {
				It("creates powershell renderer", func() {
					shellType := renderers.ShellTypePowershell
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(renderers.ShellTypePowershell))
				})
			})
			Context("WhenShellTypeIsPosix", func() {
				It("creates posix renderer", func() {
					shellType := renderers.ShellTypePosix
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(renderers.ShellTypePosix))
				})
			})
		})
		Context("WhenDarwinPlatform", func() {
			BeforeEach(func() {
				platform := "darwin"
				envGetter := &fakes.EnvGetter{Values: vars}
				factory = renderers.NewFactory(platform, envGetter)
			})
			Context("WhenDefaultShell", func() {
				Context("WhenPSModulePathEnvVarIsPresent", func() {
					It("creates powershell renderer", func() {
						vars["PSModulePath"] = "anything"
						shellType := ""
						renderer, err := factory.Create(shellType)
						Expect(err).To(BeNil())
						Expect(renderer.Type()).To(Equal(renderers.ShellTypePowershell))
						Expect(err).To(BeNil())
					})
				})
				It("creates posix renderer", func() {
					shellType := ""
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(renderers.ShellTypePosix))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shellType := renderers.ShellTypePowershell
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(renderers.ShellTypePowershell))
				})
			})
			Context("WhenShellTypeIsPosix", func() {
				It("creates posix renderer", func() {
					shellType := renderers.ShellTypePosix
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(renderers.ShellTypePosix))
				})
			})
		})
		Context("WhenLinuxPlatform", func() {
			BeforeEach(func() {
				platform := "linux"
				envGetter := &fakes.EnvGetter{Values: vars}
				factory = renderers.NewFactory(platform, envGetter)
			})
			Context("WhenDefaultShell", func() {
				Context("WhenPSModulePathEnvVarIsPresent", func() {
					It("creates powershell renderer", func() {
						vars["PSModulePath"] = "anything"
						shellType := ""
						renderer, err := factory.Create(shellType)
						Expect(err).To(BeNil())
						Expect(renderer.Type()).To(Equal(renderers.ShellTypePowershell))
						err = os.Unsetenv("PSModulePath")
						Expect(err).To(BeNil())
					})
				})
				It("creates posix renderer", func() {
					shellType := ""
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(renderers.ShellTypePosix))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shellType := renderers.ShellTypePowershell
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(shellType))
				})
			})
			Context("WhenShellTypeIsPosix", func() {
				It("creates posix renderer", func() {
					shellType := renderers.ShellTypePosix
					renderer, err := factory.Create(shellType)
					Expect(err).To(BeNil())
					Expect(renderer.Type()).To(Equal(shellType))
				})
			})
		})
	})
})
