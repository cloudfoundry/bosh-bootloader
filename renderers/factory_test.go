package renderers_test

import (
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {
	var (
		factory renderers.Factory
	)
	BeforeEach(func() {
		factory = renderers.NewFactory()
	})
	Describe("Create", func() {
		Context("WhenWindowsPlatform", func() {
			var (
				platform = "windows"
			)
			Context("WhenDefaultShell", func() {
				It("creates powershell renderer", func() {
					shell := ""
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("powershell"))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shell := "powershell"
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("powershell"))
				})
			})
			Context("WhenShellIsBash", func() {
				It("creates bash renderer", func() {
					shell := "bash"
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
		})
		Context("WhenDarwinPlatform", func() {
			var (
				platform = "darwin"
			)
			Context("WhenDefaultShell", func() {
				It("creates bash renderer", func() {
					shell := ""
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shell := "powershell"
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("powershell"))
				})
			})
			Context("WhenShellIsBash", func() {
				It("creates bash renderer", func() {
					shell := "bash"
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
		})
		Context("WhenLinuxPlatform", func() {
			var (
				platform = "linux"
			)
			Context("WhenDefaultShell", func() {
				It("creates bash renderer", func() {
					shell := ""
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal("bash"))
				})
			})
			Context("WhenShellIsPowershell", func() {
				It("creates powershell renderer", func() {
					shell := "powershell"
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal(shell))
				})
			})
			Context("WhenShellIsBash", func() {
				It("creates bash renderer", func() {
					shell := "bash"
					renderer, err := factory.Create(shell, platform)
					Expect(err).To(BeNil())
					Expect(renderer.Shell()).To(Equal(shell))
				})
			})
		})
	})
})
