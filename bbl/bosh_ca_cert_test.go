package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bosh ca cert", func() {
	It("prints CA used to sign the BOSH server cert", func() {
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
			"director-ca-cert",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("some-ca-contents"))
	})
})
