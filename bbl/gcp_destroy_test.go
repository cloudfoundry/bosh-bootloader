package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl destroy gcp", func() {
	var (
		tempDirectory string
		statePath     string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		state := storage.State{
			IAAS:    "gcp",
			TFState: "some-tf-state",
			GCP: storage.GCP{
				ProjectID:         "some-project-id",
				ServiceAccountKey: "some-service-account-key",
				Region:            "some-region",
				Zone:              "some-zone",
			},
			KeyPair: storage.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: testhelpers.BBL_KEY,
			},
			BOSH: storage.BOSH{
				DirectorName: "some-bosh-director-name",
			},
		}

		stateContents, err := json.Marshal(state)
		Expect(err).NotTo(HaveOccurred())

		statePath = filepath.Join(tempDirectory, "bbl-state.json")
		err = ioutil.WriteFile(statePath, stateContents, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	It("deletes the bbl-state", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"destroy",
		}
		cmd := exec.Command(pathToBBL, args...)

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))

		_, err = os.Stat(statePath)
		Expect(err).To(MatchError("some-error"))
	})

	It("calls out to terraform", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"destroy",
		}
		cmd := exec.Command(pathToBBL, args...)

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))

		Expect(session.Out.Contents()).To(ContainSubstring("terraform destroy"))
	})
})
