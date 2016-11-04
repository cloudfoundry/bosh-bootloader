package main_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl up", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("exits 1 and prints error message when --iaas is not provided", func() {
		session := executeCommand([]string{"--state-dir", tempDirectory, "up"}, 1)
		Expect(session.Err.Contents()).To(ContainSubstring("--iaas [gcp, aws] must be provided"))
	})

	It("exits 1 and prints error message when unsupported --iaas is provided", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "bad-iaas-value",
		}

		session := executeCommand(args, 1)
		Expect(session.Err.Contents()).To(ContainSubstring(`"bad-iaas-value" is invalid; supported values: [gcp, aws]`))
	})
})
