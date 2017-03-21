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
	"sync"

	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("bbl up gcp", func() {
	var (
		tempDirectory            string
		serviceAccountKeyPath    string
		pathToFakeBOSH           string
		pathToBOSH               string
		fakeBOSHCLIBackendServer *httptest.Server
		fakeBOSHServer           *httptest.Server
		fakeBOSH                 *fakeBOSHDirector

		fastFail                 bool
		fastFailMutex            sync.Mutex
		callRealInterpolate      bool
		callRealInterpolateMutex sync.Mutex

		createEnvArgs   string
		interpolateArgs []string
	)

	BeforeEach(func() {
		var err error
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeBOSHCLIBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/version":
				responseWriter.Write([]byte("v2.0.0"))
			case "/path":
				responseWriter.Write([]byte(originalPath))
			case "/create-env/args":
				body, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				createEnvArgs = string(body)
			case "/interpolate/args":
				body, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				interpolateArgs = append(interpolateArgs, string(body))
			case "/create-env/fastfail":
				fastFailMutex.Lock()
				defer fastFailMutex.Unlock()
				if fastFail {
					responseWriter.WriteHeader(http.StatusInternalServerError)
				} else {
					responseWriter.WriteHeader(http.StatusOK)
				}
				return
			case "/call-real-interpolate":
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				if callRealInterpolate {
					responseWriter.Write([]byte("true"))
				} else {
					responseWriter.Write([]byte("false"))
				}
			}
		}))

		fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
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
				responseWriter.Write([]byte("some-internal-tag"))
			case "/output/bosh_open_tag_name":
				responseWriter.Write([]byte("some-bosh-tag"))
			case "/version":
				responseWriter.Write([]byte("0.8.6"))
			}
		}))

		pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHCLIBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
		err = os.Rename(pathToFakeBOSH, pathToBOSH)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSH), originalPath}, ":"))

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		fastFailMutex.Lock()
		defer fastFailMutex.Unlock()
		fastFail = false
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
		Expect(state.Version).To(Equal(3))
		Expect(state.IAAS).To(Equal("gcp"))
		Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
		Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
		Expect(state.GCP.Zone).To(Equal("some-zone"))
		Expect(state.GCP.Region).To(Equal("us-west1"))
		Expect(state.KeyPair.PrivateKey).To(MatchRegexp(`-----BEGIN RSA PRIVATE KEY-----((.|\n)*)-----END RSA PRIVATE KEY-----`))
		Expect(state.KeyPair.PublicKey).To(HavePrefix("ssh-rsa"))
	})

	Context("when the terraform version is <0.8.5", func() {
		BeforeEach(func() {
			fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/version":
					responseWriter.Write([]byte("0.8.4"))
				}
			}))
		})

		It("fast fails with a helpful error message", func() {
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

			session := executeCommand(args, 1)

			Expect(session.Err.Contents()).To(ContainSubstring("Terraform version must be at least v0.8.5"))
		})
	})

	Context("when the terraform version is 0.9.0", func() {
		BeforeEach(func() {
			fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/version":
					responseWriter.Write([]byte("0.9.0"))
				}
			}))
		})

		It("fast fails with a helpful error message", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--iaas", "gcp",
				"--name", "some-bbl-env",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
			}

			session := executeCommand(args, 1)

			Expect(session.Err.Contents()).To(ContainSubstring("Version 0.9.0 of terraform is incompatible with bbl, please try a later version."))
		})
	})

	Context("when a bbl enviornment already exists", func() {
		BeforeEach(func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--iaas", "gcp",
				"--name", "some-bbl-env",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
			}

			executeCommand(args, 0)

			gcpBackend.Network.Add("some-bbl-env-network")
		})

		It("can bbl up a second time idempotently", func() {
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
		})
	})

	Context("when ops files are provides via --ops-file flag", func() {
		It("passes those ops files to bosh create env", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
				"--ops-file", "fixtures/ops-file.yml",
			}

			executeCommand(args, 0)

			Expect(interpolateArgs[0]).To(MatchRegexp(`\"-o\",\".*user-ops-file.yml\"`))
		})
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
			Expect(state.Version).To(Equal(3))
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
				Version: 3,
				IAAS:    "gcp",
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
		By("allowing the bosh interpolate call to be run", func() {
			callRealInterpolateMutex.Lock()
			defer callRealInterpolateMutex.Unlock()
			callRealInterpolate = true
		})

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

		By("resetting the ability to call interpolate to false", func() {
			callRealInterpolateMutex.Lock()
			defer callRealInterpolateMutex.Unlock()
			callRealInterpolate = false
		})
	},
		Entry("generates a cloud config with no lb type", "../cloudconfig/fixtures/gcp-cloud-config-no-lb.yml"),
	)

	Context("when there is a different environment with the same name", func() {
		var session *gexec.Session

		BeforeEach(func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
				"--name", "existing",
			}

			gcpBackend.Network.Add("existing-network")
			session = executeCommand(args, 1)
		})

		It("fast fails and prints a helpful message", func() {
			Expect(session.Err.Contents()).To(ContainSubstring("It looks like a bbl environment already exists with the name 'existing'. Please provide a different name."))
		})

		It("does not save the env id to the state", func() {
			_, err := os.Stat(filepath.Join(tempDirectory, "bbl-state.json"))
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})
	})

	Context("when the --no-director flag is provided", func() {
		It("creates the infrastructure for a bosh director", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--no-director",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
			}

			session := executeCommand(args, 0)

			Expect(session.Out.Contents()).To(ContainSubstring("terraform apply"))

		})

		It("does not invoke the bosh cli or create a cloud config", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--no-director",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
			}

			session := executeCommand(args, 0)

			Expect(session.Out.Contents()).NotTo(ContainSubstring("bosh create-env"))
			Expect(session.Out.Contents()).NotTo(ContainSubstring("step: generating cloud config"))
			Expect(session.Out.Contents()).NotTo(ContainSubstring("step: applying cloud config"))
		})
	})

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

		Context("when bosh fails", func() {
			BeforeEach(func() {
				fastFailMutex.Lock()
				fastFail = true
				fastFailMutex.Unlock()

				args := []string{
					"--state-dir", tempDirectory,
					"up",
					"--iaas", "gcp",
					"--gcp-service-account-key", serviceAccountKeyPath,
					"--gcp-project-id", "some-project-id",
					"--gcp-zone", "some-zone",
					"--gcp-region", "some-region",
				}

				executeCommand(args, 1)
			})

			It("stores a partial bosh state", func() {
				state := readStateJson(tempDirectory)
				Expect(state.BOSH.State).To(Equal(map[string]interface{}{
					"partial": "bosh-state",
				}))
			})
		})
	})
})
