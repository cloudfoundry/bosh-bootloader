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

	BeforeEach(func() {
		stdout = bytes.NewBuffer([]byte{})
		version = commands.NewVersion(stdout)
	})

	Describe("Execute", func() {
		It("prints out the version information", func() {
			err := version.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("bbl 0.0.1\n"))
		})
	})
})
