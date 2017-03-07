package main_test

import (
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl", func() {
	Context("when a v2 state exists", func() {
		var (
			tempDirectory string
		)

		BeforeEach(func() {
			var err error
			tempDirectory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			writeStateJson(storage.State{
				Version: 2,
				IAAS:    "gcp",
			}, tempDirectory)
		})

		It("fast fails with a nice error message", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
			}

			session := executeCommand(args, 1)
			Expect(session.Err.Contents()).To(ContainSubstring("Existing bbl environment is incompatible with bbl v3. Create a new environment with v3 to continue."))
		})
	})
})
