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
	var (
		tempDirectory         string
		serviceAccountKeyPath string
	)

	BeforeEach(func() {
		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes gcp details to state", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "some-region",
		}

		executeCommand(args, 0)

		state := readStateJson(tempDirectory)
		Expect(state.Version).To(Equal(2))
		Expect(state.IAAS).To(Equal("gcp"))
		Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
		Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
		Expect(state.GCP.Zone).To(Equal("some-zone"))
		Expect(state.GCP.Region).To(Equal("some-region"))
		Expect(state.KeyPair.PrivateKey).To(MatchRegexp(`-----BEGIN RSA PRIVATE KEY-----((.|\n)*)-----END RSA PRIVATE KEY-----`))
		Expect(state.KeyPair.PublicKey).To(HavePrefix("ssh-rsa"))
	})

	Context("when gcp details are provided via env vars", func() {
		BeforeEach(func() {
			os.Setenv("BBL_GCP_SERVICE_ACCOUNT_KEY", serviceAccountKeyPath)
			os.Setenv("BBL_GCP_PROJECT_ID", "some-project-id")
			os.Setenv("BBL_GCP_ZONE", "some-zone")
			os.Setenv("BBL_GCP_REGION", "some-region")
		})

		AfterEach(func() {
			os.Unsetenv("BBL_GCP_SERVICE_ACCOUNT_KEY")
			os.Unsetenv("BBL_GCP_PROJECT_ID")
			os.Unsetenv("BBL_GCP_ZONE")
			os.Unsetenv("BBL_GCP_REGION")
		})

		It("writes gcp details to state", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
				"--iaas", "gcp",
			}

			executeCommand(args, 0)

			state := readStateJson(tempDirectory)
			Expect(state.Version).To(Equal(2))
			Expect(state.IAAS).To(Equal("gcp"))
			Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
			Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
			Expect(state.GCP.Zone).To(Equal("some-zone"))
			Expect(state.GCP.Region).To(Equal("some-region"))
		})
	})

	Context("when bbl-state.json contains gcp details", func() {
		BeforeEach(func() {
			buf, err := json.Marshal(storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: `{"key": "value"}`,
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

				Expect(session.Err.Contents()).To(ContainSubstring("The iaas type cannot be changed for an existing environment. The current iaas type is gcp."))
			})
		})
	})
})
