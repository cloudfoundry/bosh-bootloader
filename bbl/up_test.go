package main_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl up", func() {
	var (
		tempDirectory         string
		serviceAccountKeyPath string
		fakeBOSHServer        *httptest.Server
		fakeBOSH              *fakeBOSHDirector
	)

	BeforeEach(func() {
		var err error
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/output/external_ip":
				responseWriter.Write([]byte("127.0.0.1"))
			case "/output/director_address":
				responseWriter.Write([]byte(fakeBOSHServer.URL))
			case "/version":
				responseWriter.Write([]byte("0.8.6"))
			}
		}))

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		fakeBOSHCLIBackendServer.ResetAll()
	})

	It("writes iaas to state", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-west1",
		}

		executeCommand(args, 0)

		state := readStateJson(tempDirectory)
		Expect(state.IAAS).To(Equal("gcp"))
	})

	It("succeeds if provided an empty json struct as the bbl-state.json", func() {
		err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), []byte("{}"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-west1",
		}

		executeCommand(args, 0)
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
				"--gcp-region", "us-west1",
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
		Expect(session.Err.Contents()).To(ContainSubstring(`"bad-iaas-value" is an invalid iaas type, supported values are: [gcp, aws]`))
	})

	Context("when the bosh cli version is <2.0", func() {
		BeforeEach(func() {
			fakeBOSHCLIBackendServer.SetVersion("1.9.0")
		})

		It("fast fails with a helpful error message", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
			}

			session := executeCommand(args, 1)

			Expect(session.Err.Contents()).To(ContainSubstring("BOSH version must be at least v2.0.0"))
		})
	})
})
