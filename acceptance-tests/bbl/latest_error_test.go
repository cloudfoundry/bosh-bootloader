package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl latest-error", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		acceptance.SkipUnless("latest-error")

		var err error
		tempDirectory, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		state := []byte(`{
			"version": 8,
			"noDirector": true,
			"tfState": "some-tf-state",
			"latestTFOutput": "some terraform output"
		}`)
		err = os.WriteFile(filepath.Join(tempDirectory, storage.STATE_FILE), state, storage.StateMode)
		Expect(err).NotTo(HaveOccurred())
	})

	It("prints the terraform output from the last command", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"latest-error",
		}

		cmd := exec.Command(pathToBBL, args...)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, bblLatestErrorTimeout).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(ContainSubstring("some terraform output"))
	})
})
