package acceptance_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl latest-error", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		state := []byte(`{
			"version": 8,
			"noDirector": true,
			"tfState": "some-tf-state",
			"latestTFOutput": "some terraform output"
		}`)
		err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	It("prints the terraform output from the last command", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"latest-error",
		}

		session := executeCommand(args, 0)

		Expect(session.Out.Contents()).To(ContainSubstring("some terraform output"))
	})
})
