package renderers_test

import (
	"os"

	"github.com/cloudfoundry/bosh-bootloader/renderers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {
	BeforeEach(func() {
		os.Unsetenv("PSModulePath")
	})
	Describe("Create", func() {
		Context("WhenWindowsPlatform", func() {
			var (
				platform = "windows"
				factory  renderers.Factory
			)
			BeforeEach(func() {
				factory = renderers.NewFactory(platform)
			})
			Context("WhenDefaultShell", func() {
				It("creates powershell renderer", func() {
					shell := ""
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("powershell"))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shell := "powershell"
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("powershell"))
				})
			})
			Context("WhenShellIsBash", func() {
				It("creates bash renderer", func() {
					shell := "bash"
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
		})
		Context("WhenDarwinPlatform", func() {
			var (
				platform = "darwin"
				factory  renderers.Factory
			)
			BeforeEach(func() {
				factory = renderers.NewFactory(platform)
			})
			Context("WhenDefaultShell", func() {
				Context("WhenPSModulePathEnvVarIsPresent", func() {
					It("creates powershell renderer", func() {
						err := os.Setenv("PSModulePath", "anything")
						Expect(err).To(BeNil())
						shell := ""
						renderer, err := factory.Create(shell)
						Expect(err).To(BeNil())
						Expect(renderer.Shell()).To(Equal("powershell"))
						Expect(err).To(BeNil())
					})
				})
				It("creates bash renderer", func() {
					shell := ""
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shell := "powershell"
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("powershell"))
				})
			})
			Context("WhenShellIsBash", func() {
				It("creates bash renderer", func() {
					shell := "bash"
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
		})
		Context("WhenLinuxPlatform", func() {
			var (
				platform = "linux"
				factory  renderers.Factory
			)
			BeforeEach(func() {
				factory = renderers.NewFactory(platform)
			})
			Context("WhenDefaultShell", func() {
				Context("WhenPSModulePathEnvVarIsPresent", func() {
					It("creates powershell renderer", func() {
						err := os.Setenv("PSModulePath", "anything")
						Expect(err).To(BeNil())
						shell := ""
						renderer, err := factory.Create(shell)
						Expect(err).To(BeNil())
						Expect(renderer.Shell()).To(Equal("powershell"))
						err = os.Unsetenv("PSModulePath")
						Expect(err).To(BeNil())
					})
				})
				It("creates bash renderer", func() {
					shell := ""
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shell := "powershell"
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal(shell))
				})
			})
			Context("WhenShellIsBash", func() {
				It("creates bash renderer", func() {
					shell := "bash"
					renderer, err := factory.Create(shell)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal(shell))
				})
			})
		})
	})
})
