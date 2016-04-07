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

var _ = Describe("director-password", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns the director password from the given state file", func() {
		state := []byte(`{
			"bosh": {
				"directorPassword": "some-director-password"
			}
		}`)
		err := ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), state, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"director-password",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("some-director-password"))
	})

	Context("failure cases", func() {
		It("returns a non zero exit code when the password does not exist", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"director-password",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring("Could not retrieve director password"))
		})
	})
})
