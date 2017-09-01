package acceptance_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
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
		acceptance.SkipUnless("latest-error")
		args := []string{
			"--state-dir", tempDirectory,
			"latest-error",
		}

		session := executeCommand(args, 0)

		Expect(string(session.Out.Contents())).To(ContainSubstring("some terraform output"))
	})
})

func executeCommand(args []string, exitCode int) *gexec.Session {
	cmd := exec.Command(pathToBBL, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(exitCode))

	return session
}
