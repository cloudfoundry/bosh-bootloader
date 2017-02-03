package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("bbl up gcp", func() {
	var (
		tempDirectory              string
		serviceAccountKeyPath      string
		pathToFakeTerraform        string
		pathToTerraform            string
		pathToFakeBOSH             string
		pathToBOSH                 string
		fakeBOSHCLIBackendServer   *httptest.Server
		fakeTerraformBackendServer *httptest.Server
		fakeBOSHServer             *httptest.Server
		fakeBOSH                   *fakeBOSHDirector
	)

	BeforeEach(func() {
		var err error
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeBOSHCLIBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		}))

		fakeTerraformBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/output/external_ip":
				responseWriter.Write([]byte("127.0.0.1"))
			case "/output/director_address":
				responseWriter.Write([]byte(fakeBOSHServer.URL))
			case "/output/network_name":
				responseWriter.Write([]byte("some-network-name"))
			case "/output/subnetwork_name":
				responseWriter.Write([]byte("some-subnetwork-name"))
			case "/output/internal_tag_name":
				responseWriter.Write([]byte("some-tag"))
			case "/output/bosh_open_tag_name":
				responseWriter.Write([]byte("some-bosh-open-tag"))
			}
		}))

		pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
		err = os.Rename(pathToFakeTerraform, pathToTerraform)
		Expect(err).NotTo(HaveOccurred())

		pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHCLIBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
		err = os.Rename(pathToFakeBOSH, pathToBOSH)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), filepath.Dir(pathToBOSH), originalPath}, ":"))

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
	})

	It("writes gcp details to state", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"--debug",
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-west1",
		}

		executeCommand(args, 0)

		state := readStateJson(tempDirectory)
		Expect(state.Version).To(Equal(2))
		Expect(state.IAAS).To(Equal("gcp"))
		Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
		Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
		Expect(state.GCP.Zone).To(Equal("some-zone"))
		Expect(state.GCP.Region).To(Equal("us-west1"))
		Expect(state.KeyPair.PrivateKey).To(MatchRegexp(`-----BEGIN RSA PRIVATE KEY-----((.|\n)*)-----END RSA PRIVATE KEY-----`))
		Expect(state.KeyPair.PublicKey).To(HavePrefix("ssh-rsa"))
	})

	Context("when gcp details are provided via env vars", func() {
		BeforeEach(func() {
			os.Setenv("BBL_GCP_SERVICE_ACCOUNT_KEY", serviceAccountKeyPath)
			os.Setenv("BBL_GCP_PROJECT_ID", "some-project-id")
			os.Setenv("BBL_GCP_ZONE", "some-zone")
			os.Setenv("BBL_GCP_REGION", "us-west1")
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
			Expect(state.GCP.Region).To(Equal("us-west1"))
		})
	})

	Context("when bbl-state.json contains gcp details", func() {
		BeforeEach(func() {
			buf, err := json.Marshal(storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: serviceAccountKey,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
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

		Context("when re-bbling up with different gcp args than in bbl-state", func() {
			It("returns an error when passing different region", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"up",
					"--iaas", "gcp",
					"--gcp-region", "some-other-region",
					"--gcp-zone", "some-zone",
					"--gcp-project-id", "some-project-id",
					"--gcp-service-account-key", serviceAccountKeyPath,
				}
				session := executeCommand(args, 1)
				Expect(session.Err.Contents()).To(ContainSubstring("The region cannot be changed for an existing environment. The current region is us-west1."))
			})

			It("returns an error when passing different zone", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"up",
					"--iaas", "gcp",
					"--gcp-zone", "some-other-zone",
					"--gcp-region", "us-west1",
					"--gcp-project-id", "some-project-id",
					"--gcp-service-account-key", serviceAccountKeyPath,
				}
				session := executeCommand(args, 1)
				Expect(session.Err.Contents()).To(ContainSubstring("The zone cannot be changed for an existing environment. The current zone is some-zone."))
			})

			It("returns an error when passing different project-id", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"up",
					"--iaas", "gcp",
					"--gcp-project-id", "some-other-project-id",
					"--gcp-zone", "some-zone",
					"--gcp-region", "us-west1",
					"--gcp-service-account-key", serviceAccountKeyPath,
				}
				session := executeCommand(args, 1)
				Expect(session.Err.Contents()).To(ContainSubstring("The project id cannot be changed for an existing environment. The current project id is some-project-id."))
			})
		})
	})

	It("calls out to terraform", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"--debug",
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-west1",
		}

		session := executeCommand(args, 0)

		Expect(session.Out.Contents()).To(ContainSubstring("terraform apply"))
	})

	It("invokes the bosh cli", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"--debug",
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-west1",
		}

		session := executeCommand(args, 0)

		Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))
	})

	It("can invoke the bosh cli idempotently", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"--debug",
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-west1",
		}

		session := executeCommand(args, 0)
		Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))

		session = executeCommand(args, 0)
		Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))
		Expect(session.Out.Contents()).To(ContainSubstring("No new changes, skipping deployment..."))
	})

	DescribeTable("cloud config", func(fixtureLocation string) {
		contents, err := ioutil.ReadFile(fixtureLocation)
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-east1",
		}

		session := executeCommand(args, 0)
		stdout := session.Out.Contents()

		Expect(stdout).To(ContainSubstring("step: generating cloud config"))
		Expect(stdout).To(ContainSubstring("step: applying cloud config"))
		Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))

		By("executing idempotently", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
			}

			executeCommand(args, 0)
			Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))
		})
	},
		Entry("generates a cloud config with no lb type", "../cloudconfig/gcp/fixtures/cloud-config-no-lb.yml"),
	)

	Context("bbl re-entrance", func() {
		It("saves the tf state when terraform apply fails", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "fail-to-terraform",
			}

			executeCommand(args, 1)

			state := readStateJson(tempDirectory)
			Expect(state.TFState).To(Equal(`{"key":"partial-apply"}`))
		})
	})
})
