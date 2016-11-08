package main_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl up", func() {
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
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(`{"real": "json"}`), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes iaas to state", func() {
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
		Expect(state.IAAS).To(Equal("gcp"))
	})

	Context("when providing iaas via env vars", func() {
		BeforeEach(func() {
			err := os.Setenv("BBL_IAAS", "gcp")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Unsetenv("BBL_IAAS")
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes iaas to state", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "some-region",
			}

			executeCommand(args, 0)

			state := readStateJson(tempDirectory)
			Expect(state.IAAS).To(Equal("gcp"))
		})
	})

	It("exits 1 and prints error message when --iaas is not provided", func() {
		session := executeCommand([]string{"--state-dir", tempDirectory, "up"}, 1)
		Expect(session.Err.Contents()).To(ContainSubstring("--iaas [gcp, aws] must be provided"))
	})

	It("exits 1 and prints error message when unsupported --iaas is provided", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "bad-iaas-value",
		}

		session := executeCommand(args, 1)
		Expect(session.Err.Contents()).To(ContainSubstring(`"bad-iaas-value" is invalid; supported values: [gcp, aws]`))
	})
})
