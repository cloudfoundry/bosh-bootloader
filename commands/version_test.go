package commands_test

import (
	"bytes"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var (
		version commands.Version
		stdout  *bytes.Buffer
	)

	Context("when no version number was passed in", func() {
		BeforeEach(func() {
			stdout = bytes.NewBuffer([]byte{})
			version = commands.NewVersion("", stdout)
		})

		Describe("Execute", func() {
			It("prints out dev as the version", func() {
				err := version.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout.String()).To(Equal("bbl dev\n"))
			})
		})
	})

	Context("when a version number was passed in", func() {
		BeforeEach(func() {
			stdout = bytes.NewBuffer([]byte{})
			version = commands.NewVersion("1.2.3", stdout)
		})

		Describe("Execute", func() {
			It("prints out the passed in version information", func() {
				err := version.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout.String()).To(Equal("bbl 1.2.3\n"))
			})
		})
	})
})
