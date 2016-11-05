package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl up gcp", func() {
	var tempDirectory string

	BeforeEach(func() {
		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes gcp details to state", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", "some-service-account-key",
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "some-region",
		}

		executeCommand(args, 0)

		state := readStateJson(tempDirectory)
		Expect(state).To(Equal(storage.State{
			Version: 2,
			IAAS:    "gcp",
			GCP: storage.GCP{
				ServiceAccountKey: "some-service-account-key",
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "some-region",
			},
		}))
	})

	Context("when bbl-state.json contains gcp details", func() {
		BeforeEach(func() {
			buf, err := json.Marshal(storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not require gcp args and exits 0", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
			}

			executeCommand(args, 0)
		})

		Context("when called with --iaas aws", func() {
			It("exits 1 and prints error message", func() {
				session := upAWS("", tempDirectory, 1)

				Expect(session.Err.Contents()).To(ContainSubstring("the iaas provided must match the iaas in bbl-state.json"))
			})
		})
	})
})
