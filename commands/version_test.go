package commands_test

import (
	"fmt"
	"runtime"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var (
		version commands.Version
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
	})

	Describe("CheckFastFails", func() {
		BeforeEach(func() {
			version = commands.NewVersion("dev", logger)
		})

		It("returns no error", func() {
			err := version.CheckFastFails([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Execute", func() {
		Context("when no version number was passed in", func() {
			BeforeEach(func() {
				version = commands.NewVersion("dev", logger)
			})

			Describe("Execute", func() {
				It("prints out dev as the version", func() {
					err := version.Execute([]string{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
						fmt.Sprintf("bbl dev (%s/%s)\n", runtime.GOOS, runtime.GOARCH),
					}))
				})
			})
		})

		Context("when a version number was passed in", func() {
			BeforeEach(func() {
				version = commands.NewVersion("1.2.3", logger)
			})

			Describe("Execute", func() {
				It("prints out the passed in version information", func() {
					err := version.Execute([]string{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
						fmt.Sprintf("bbl 1.2.3 (%s/%s)\n", runtime.GOOS, runtime.GOARCH),
					}))
				})
			})
		})
	})
})
