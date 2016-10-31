package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("director-ca-cert", func() {
	DescribeTable("prints CA used to sign the BOSH server cert",
		func(command string) {
			tempDirectory, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			state := []byte(`{
				"bosh": {
					"directorSSLCA": "some-ca-contents"
				}
			}`)
			err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				command,
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("some-ca-contents"))
		},
		Entry("director-ca-cert", "director-ca-cert"),
		Entry("supporting bosh-ca-cert for backwards compatibility", "bosh-ca-cert"),
	)

	It("returns a non zero exit code when the bbl-state.json does not exist", func() {
		tempDirectory, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"director-ca-cert",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))

		expectedErrorMessage := fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)
		Expect(session.Err.Contents()).To(ContainSubstring(expectedErrorMessage))
	})

	Context("bosh-ca-cert", func() {
		It("returns a non zero exit code when the bbl-state.json does not exist", func() {
			tempDirectory, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"bosh-ca-cert",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))

			expectedErrorMessage := fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)
			Expect(session.Err.Contents()).To(ContainSubstring(expectedErrorMessage))
		})

		It("prints a deprecation warning to STDERR", func() {
			tempDirectory, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			state := []byte(`{}`)
			err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(exec.Command(pathToBBL, "--state-dir", tempDirectory, "bosh-ca-cert"), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit())
			Expect(session.Err.Contents()).To(ContainSubstring("'bosh-ca-cert' has been deprecated and will be removed in future versions of bbl, please use 'director-ca-cert'"))
		})
	})
})
