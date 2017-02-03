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

var _ = Describe("bbl", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("prints the bosh environment variables when print-env is called to be sourced", func() {
		state := []byte(`{
			"bosh": {
				"directorAddress": "some-director-address",
				"directorUsername": "some-director-username",
				"directorPassword": "some-director-password",
				"directorSSLCA": "some-director-ca-cert"
			},
			"envID": "some-env-id"
		}`)
		err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"print-env",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_CLIENT=some-director-username"))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_CLIENT_SECRET=some-director-password"))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_CA_CERT='some-director-ca-cert'"))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_ENVIRONMENT=some-director-address"))
	})

})
