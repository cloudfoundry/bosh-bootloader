package main_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bbl/fakejumpbox"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl up gcp", func() {
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

		fakeTerraformBackendServer.SetFakeBOSHServer(fakeBOSHServer.URL)

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
		fakeTerraformBackendServer.ResetAll()
	})

	It("creates infrastructure on GCP", func() {
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

		By("writing gcp details to state", func() {
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

		By("writing logging messages to stdout", func() {
			stdout := session.Out.Contents()

			Expect(stdout).To(ContainSubstring("step: appending new ssh-keys"))
			Expect(stdout).To(ContainSubstring("step: generating terraform template"))
			Expect(stdout).To(ContainSubstring("step: applied terraform template"))
			Expect(stdout).To(ContainSubstring("step: creating bosh director"))
			Expect(stdout).To(ContainSubstring("step: created bosh director"))
			Expect(stdout).To(ContainSubstring("step: generating cloud config"))
			Expect(stdout).To(ContainSubstring("step: applying cloud config"))
		})

		By("calling out to terraform", func() {
			Expect(session.Out.Contents()).To(ContainSubstring("terraform apply"))
		})

		By("invoking the bosh cli", func() {
			Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))
		})
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

	Context("when the gcp service account key not passed as a file", func() {
		It("accepts the service account key contents", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKey,
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
	})

	Context("when provided a name with invalid characters", func() {
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
				"--name", "some_name",
			}

			session := executeCommand(args, 1)

			Expect(session.Err.Contents()).To(ContainSubstring("Names must start with a letter and be alphanumeric or hyphenated."))
		})
	})

	Context("when the terraform version is <0.8.5", func() {
		BeforeEach(func() {
			fakeTerraformBackendServer.SetVersion("0.8.4")
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
			fakeTerraformBackendServer.SetVersion("0.9.0")
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

	Context("when a bbl environment already exists", func() {
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

	Context("when a user provides an ops file via the --ops-file flag", func() {
		BeforeEach(func() {
			fakeBOSHCLIBackendServer.SetCallRealInterpolate(true)
		})

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

			Expect(fakeBOSHCLIBackendServer.GetInterpolateArgs(1)).To(MatchRegexp(`\"-o\",\".*user-ops-file.yml\"`))
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

	DescribeTable("cloud config", func(fixtureLocation string) {
		By("allowing the bosh interpolate call to be run", func() {
			fakeBOSHCLIBackendServer.SetCallRealInterpolate(true)
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
			"--gcp-region", "us-west1",
		}

		executeCommand(args, 0)

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

	Context("when the --jumpbox flag is provided", func() {
		var (
			jumpboxServer       *fakejumpbox.JumpboxServer
			fakeHTTPSBOSHServer *httptest.Server
		)

		BeforeEach(func() {
			fakeHTTPSBOSHServer = httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				fakeBOSH.ServeHTTP(responseWriter, request)
			}))
			jumpboxServer = fakejumpbox.NewJumpboxServer()
			jumpboxServer.Start(testhelpers.JUMPBOX_SSH_KEY, fakeHTTPSBOSHServer.Listener.Addr().String())

			fakeTerraformBackendServer.SetJumpboxURLOutput(jumpboxServer.Addr())
		})

		It("creates a jumpbox using bosh create-env", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--iaas", "gcp",
				"--jumpbox",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
			}

			session := executeCommand(args, 0)

			Expect(fakeBOSHCLIBackendServer.CreateEnvCallCount()).To(Equal(2))
			Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))
			Expect(session.Out.Contents()).To(ContainSubstring("step: creating jumpbox"))
			Expect(session.Out.Contents()).To(ContainSubstring("step: created jumpbox"))
			Expect(session.Out.Contents()).To(ContainSubstring("step: creating bosh director"))
			Expect(session.Out.Contents()).To(ContainSubstring("step: created bosh director"))
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
		Context("when terraform apply fails", func() {
			var (
				session *gexec.Session
			)

			BeforeEach(func() {
				args := []string{
					"--state-dir", tempDirectory,
					"up",
					"--iaas", "gcp",
					"--gcp-service-account-key", serviceAccountKeyPath,
					"--gcp-project-id", "some-project-id",
					"--gcp-zone", "some-zone",
					"--gcp-region", "fail-to-terraform",
				}

				session = executeCommand(args, 1)
			})

			It("saves the tf state", func() {
				state := readStateJson(tempDirectory)
				Expect(state.TFState).To(Equal(`{"key":"partial-apply"}`))
			})

			Context("when no --debug flag is provided", func() {
				It("returns a helpful error message", func() {
					Expect(session.Err.Contents()).To(ContainSubstring("Some output has been redacted, use `bbl latest-error` to see it or run again with --debug for additional debug output"))
				})
			})
		})

		Context("when bosh fails", func() {
			BeforeEach(func() {
				fakeBOSHCLIBackendServer.SetCreateEnvFastFail(true)

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
