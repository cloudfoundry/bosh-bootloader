package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("director-address", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns the director address from the given state file", func() {
		state := []byte(`{
			"bosh": {
				"directorAddress": "some-director-url"
			}
		}`)
		err := ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), state, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"director-address",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("some-director-url"))
	})

	Context("failure cases", func() {
		It("returns a non zero exit code when the address does not exist", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"director-address",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring("Could not retrieve director address"))
		})
	})
})
