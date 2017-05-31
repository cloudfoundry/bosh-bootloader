package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	compute "google.golang.org/api/compute/v1"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const (
	variablesYAML = `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`
)

var _ = Describe("GCPUp", func() {
	var (
		gcpUp                 commands.GCPUp
		stateStore            *fakes.StateStore
		keyPairManager        *fakes.KeyPairManager
		gcpClientProvider     *fakes.GCPClientProvider
		gcpClient             *fakes.GCPClient
		terraformManager      *fakes.TerraformManager
		boshManager           *fakes.BOSHManager
		cloudConfigManager    *fakes.CloudConfigManager
		envIDManager          *fakes.EnvIDManager
		logger                *fakes.Logger
		terraformManagerError *fakes.TerraformManagerError

		serviceAccountKeyPath string
		serviceAccountKey     string

		expectedIAASState      storage.State
		expectedEnvIDState     storage.State
		expectedKeyPairState   storage.State
		expectedTerraformState storage.State
		expectedBOSHState      storage.State

		expectedTerraformTemplate string
	)

	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		keyPairManager = &fakes.KeyPairManager{}
		gcpClientProvider = &fakes.GCPClientProvider{}
		gcpClient = &fakes.GCPClient{}
		gcpClientProvider.ClientCall.Returns.Client = gcpClient
		gcpClient.GetNetworksCall.Returns.NetworkList = &compute.NetworkList{}
		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}
		terraformManager = &fakes.TerraformManager{}
		envIDManager = &fakes.EnvIDManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		terraformManagerError = &fakes.TerraformManagerError{}

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		serviceAccountKey = `{"real": "json"}`
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		expectedIAASState = storage.State{
			IAAS: "gcp",
			GCP: storage.GCP{
				ServiceAccountKey: serviceAccountKey,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			},
		}

		expectedEnvIDState = expectedIAASState
		expectedEnvIDState.EnvID = "some-env-id"

		expectedKeyPairState = expectedEnvIDState
		expectedKeyPairState.KeyPair = storage.KeyPair{
			PrivateKey: "some-private-key",
			PublicKey:  "some-public-key",
		}

		expectedTerraformState = expectedKeyPairState
		expectedTerraformState.TFState = "some-tf-state"

		expectedBOSHState = expectedTerraformState
		expectedBOSHState.BOSH = storage.BOSH{
			DirectorName:           "bosh-some-env-id",
			DirectorUsername:       "admin",
			DirectorPassword:       "some-admin-password",
			DirectorAddress:        "some-director-address",
			DirectorSSLCA:          "some-ca",
			DirectorSSLCertificate: "some-certificate",
			DirectorSSLPrivateKey:  "some-private-key",
			State: map[string]interface{}{
				"new-key": "new-value",
			},
			Variables: variablesYAML,
			Manifest:  "some-bosh-manifest",
		}

		terraformManager.VersionCall.Returns.Version = "0.8.7"
		envIDManager.SyncCall.Returns.State = storage.State{
			EnvID: "some-env-id",
		}
		keyPairManager.SyncCall.Returns.State = expectedKeyPairState
		terraformManager.ApplyCall.Returns.BBLState = expectedTerraformState
		boshManager.CreateCall.Returns.State = expectedBOSHState

		gcpUp = commands.NewGCPUp(commands.NewGCPUpArgs{
			StateStore:         stateStore,
			KeyPairManager:     keyPairManager,
			GCPProvider:        gcpClientProvider,
			TerraformManager:   terraformManager,
			BoshManager:        boshManager,
			Logger:             logger,
			EnvIDManager:       envIDManager,
			CloudConfigManager: cloudConfigManager,
		})

		body, err := ioutil.ReadFile("fixtures/terraform_template_no_lb.tf")
		Expect(err).NotTo(HaveOccurred())

		expectedTerraformTemplate = string(body)
	})

	AfterEach(func() {
		commands.ResetMarshal()
	})

	Describe("Execute", func() {
		It("sets the GCP configuration", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(gcpClientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(gcpClientProvider.SetConfigCall.Receives.ServiceAccountKey).To(Equal(serviceAccountKey))
			Expect(gcpClientProvider.SetConfigCall.Receives.ProjectID).To(Equal("some-project-id"))
			Expect(gcpClientProvider.SetConfigCall.Receives.Zone).To(Equal("some-zone"))
		})

		It("sets the serviceAccountKey from the path", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(gcpClientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(gcpClientProvider.SetConfigCall.Receives.ServiceAccountKey).To(Equal(serviceAccountKey))
			Expect(gcpClientProvider.SetConfigCall.Receives.ProjectID).To(Equal("some-project-id"))
			Expect(gcpClientProvider.SetConfigCall.Receives.Zone).To(Equal("some-zone"))
		})

		It("sets the serviceAccountKey from the given JSON string", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKey,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(gcpClientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(gcpClientProvider.SetConfigCall.Receives.ServiceAccountKey).To(Equal(serviceAccountKey))
			Expect(gcpClientProvider.SetConfigCall.Receives.ProjectID).To(Equal("some-project-id"))
			Expect(gcpClientProvider.SetConfigCall.Receives.Zone).To(Equal("some-zone"))
		})

		It("retrieves the env ID", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(envIDManager.SyncCall.Receives.State).To(Equal(expectedIAASState))
			Expect(envIDManager.SyncCall.Receives.Name).To(BeEmpty())
		})

		It("saves the resulting state with the env ID", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(expectedEnvIDState))
		})

		It("syncs the keypair", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					ServiceAccountKey: `{"real": "json"}`,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairManager.SyncCall.CallCount).To(Equal(1))
			Expect(keyPairManager.SyncCall.Receives.State).To(Equal(storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					ServiceAccountKey: `{"real": "json"}`,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				},
			}))
		})

		It("saves the key pair to the state", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 2))
			Expect(stateStore.SetCall.Receives[1].State).To(Equal(expectedKeyPairState))
		})

		It("creates gcp resources via terraform", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(expectedKeyPairState))
		})

		It("saves the terraform state to the state", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 3))
			Expect(stateStore.SetCall.Receives[2].State).To(Equal(expectedTerraformState))
		})

		It("creates a bosh", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
			Expect(boshManager.CreateCall.Receives.State).To(Equal(expectedTerraformState))
		})

		It("saves the bosh state to the state", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 4))
			Expect(stateStore.SetCall.Receives[3].State).To(Equal(expectedBOSHState))
		})

		It("updates the cloud config", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "some-region",
			}, storage.State{
				EnvID: "bbl-lake-time:stamp",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(expectedBOSHState))
		})

		Context("when a name is passed in for env-id", func() {
			It("passes that name in for the env id manager to use", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
					Name:              "some-other-env-id",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
				Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-other-env-id"))
			})
		})

		Context("when ops file are passed in", func() {
			It("passes the ops file contents to the bosh manager", func() {
				opsFile, err := ioutil.TempFile("", "ops-file")
				Expect(err).NotTo(HaveOccurred())

				opsFilePath := opsFile.Name()
				opsFileContents := "some-ops-file-contents"
				err = ioutil.WriteFile(opsFilePath, []byte(opsFileContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
					OpsFilePath:       opsFilePath,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateCall.Receives.State.BOSH.UserOpsFile).To(Equal("some-ops-file-contents"))
			})
		})

		Context("when the no-director flag is provided", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true
			})

			It("does not create a bosh or update cloud config", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
					NoDirector:        true,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateCall.CallCount).To(Equal(0))
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
				Expect(stateStore.SetCall.CallCount).To(Equal(3))
				Expect(stateStore.SetCall.Receives[2].State.NoDirector).To(Equal(true))
			})

			Context("when re-bbling up an environment with no director", func() {
				It("does not create a bosh director", func() {
					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					}, storage.State{
						NoDirector: true,
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
					Expect(boshManager.CreateCall.CallCount).To(Equal(0))
					Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State.NoDirector).To(Equal(true))
				})
			})
		})

		Context("when the jumpbox flag is provided", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.Jumpbox.Enabled = true
			})

			It("creates a jumpbox", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
					NoDirector:        false,
					Jumpbox:           true,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateCall.CallCount).To(Equal(1))
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.CallCount).To(Equal(4))
				Expect(stateStore.SetCall.Receives[0].State.Jumpbox.Enabled).To(Equal(true))
			})
		})

		Context("reentrance", func() {
			var (
				updatedServiceAccountKey     string
				updatedServiceAccountKeyPath string
			)

			BeforeEach(func() {
				tempFile, err := ioutil.TempFile("", "updatedGcpServiceAccountKey")
				Expect(err).NotTo(HaveOccurred())

				updatedServiceAccountKeyPath = tempFile.Name()
				updatedServiceAccountKey = `{"another-real": "json-file"}`
				err = ioutil.WriteFile(updatedServiceAccountKeyPath, []byte(updatedServiceAccountKey), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not require details from up config", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not store the state if the provided flags are not valid", func() {
				err := gcpUp.Execute(
					commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
					}, storage.State{})
				Expect(err).To(MatchError("GCP project ID must be provided"))
				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			It("saves the keypair when the terraform fails", func() {
				terraformManager.ApplyCall.Returns.Error = errors.New("terraform manager failed")

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("terraform manager failed"))

				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[1].State).To(Equal(expectedKeyPairState))
			})

			It("calls terraform manager with previous state", func() {
				expectedKeyPairState.TFState = "existing-tf-state"
				err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					},
					TFState: "existing-tf-state",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(expectedKeyPairState))
			})
		})

		Context("failure cases", func() {
			It("returns an error if terraform manager version validator fails", func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("cannot validate version")

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})

				Expect(err).To(MatchError("cannot validate version"))
			})

			It("returns an error when the service account key passed in is not an existent filename or valid json", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: "/some/non/existent/file",
					ProjectID:         "p",
					Zone:              "z",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("error reading or parsing service account key (must be valid json or a file containing valid json): invalid character '/' looking for beginning of value"))
			})

			It("returns an error when the service account key file does not contain valid json", func() {
				tempFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				invalidServiceAccountKeyPath := tempFile.Name()
				err = ioutil.WriteFile(invalidServiceAccountKeyPath, []byte(`%%%not-valid-json%%%`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: invalidServiceAccountKeyPath,
					ProjectID:         "p",
					Zone:              "z",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("error reading or parsing service account key (must be valid json or a file containing valid json): invalid character '%' looking for beginning of value"))
			})

			It("returns an error when the ops file cannot be read", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					OpsFilePath:       "some/fake/path",
				}, storage.State{})
				Expect(err).To(MatchError("error reading ops-file contents: open some/fake/path: no such file or directory"))
			})

			Context("when calling up with different gcp flags then the state", func() {
				It("returns an error when the --gcp-region is different", func() {
					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-other-region",
					}, storage.State{
						GCP: storage.GCP{
							ServiceAccountKey: serviceAccountKey,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "us-west1",
						},
					})
					Expect(err).To(MatchError("The region cannot be changed for an existing environment. The current region is us-west1."))
				})

				It("returns an error when the --gcp-zone is different", func() {
					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-project-id",
						Zone:              "some-other-zone",
						Region:            "us-west1",
					}, storage.State{
						GCP: storage.GCP{
							ServiceAccountKey: serviceAccountKey,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "us-west1",
						},
					})
					Expect(err).To(MatchError("The zone cannot be changed for an existing environment. The current zone is some-zone."))
				})

				It("returns an error when the --gcp-project-id is different", func() {
					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-other-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					}, storage.State{
						GCP: storage.GCP{
							ServiceAccountKey: serviceAccountKey,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "us-west1",
						},
					})
					Expect(err).To(MatchError("The project id cannot be changed for an existing environment. The current project id is some-project-id."))
				})
			})

			Context("when a bbl environment exists with a bosh director", func() {
				It("fast fails before creating any infrastructure", func() {
					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
						NoDirector:        true,
					}, storage.State{
						BOSH: storage.BOSH{
							DirectorName: "some-director",
						},
					})
					Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))

					Expect(envIDManager.SyncCall.CallCount).To(Equal(0))
					Expect(keyPairManager.SyncCall.CallCount).To(Equal(0))
					Expect(terraformManager.ApplyCall.CallCount).To(Equal(0))
					Expect(boshManager.CreateCall.CallCount).To(Equal(0))
				})
			})

			DescribeTable("up config validation", func(upConfig func() commands.GCPUpConfig, expectedErr string) {
				err := gcpUp.Execute(upConfig(), storage.State{})
				Expect(err).To(MatchError(expectedErr))
			},
				Entry("returns an error when no flags are passed in", func() commands.GCPUpConfig {
					return commands.GCPUpConfig{}
				},
					"GCP service account key must be provided"),
				Entry("returns an error when service account key is missing", func() commands.GCPUpConfig {
					return commands.GCPUpConfig{
						ProjectID: "p",
						Zone:      "z",
						Region:    "us-west1",
					}
				}, "GCP service account key must be provided"),
				Entry("returns an error when project ID is missing", func() commands.GCPUpConfig {
					return commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						Zone:              "z",
						Region:            "us-west1",
					}
				}, "GCP project ID must be provided"),
				Entry("returns an error when zone is missing", func() commands.GCPUpConfig {
					return commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "p",
						Region:            "us-west1",
					}
				}, "GCP zone must be provided"),
				Entry("returns an error when region is missing", func() commands.GCPUpConfig {
					return commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "p",
						Zone:              "z",
					}
				}, "GCP region must be provided"),
			)

			It("returns an error when setting config fails", func() {
				gcpClientProvider.SetConfigCall.Returns.Error = errors.New("setting config failed")

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("setting config failed"))
			})

			It("fast fails if a gcp environment with the same name already exists", func() {
				envIDManager.SyncCall.Returns.Error = errors.New("environment already exists")
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})

				Expect(err).To(MatchError("environment already exists"))
			})

			It("returns an error when state store fails to set after syncing env id", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("set call failed")}}
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "p",
					Zone:              "z",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("set call failed"))
			})

			It("returns an error when the keypair could not be updated", func() {
				keyPairManager.SyncCall.Returns.Error = errors.New("keypair sync failed")

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("keypair sync failed"))
			})

			It("returns an error when the state fails to be set after updating keypair", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("state failed to be set")}}

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("state failed to be set"))
			})

			Context("terraform manager error handling", func() {
				BeforeEach(func() {
					terraformManagerError.ErrorCall.Returns = "failed to apply"
					terraformManagerError.BBLStateCall.Returns.BBLState = storage.State{
						TFState: "some-updated-tf-state",
					}
				})

				It("saves the tf state when the applier fails", func() {
					terraformManager.ApplyCall.Returns.Error = terraformManagerError

					err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
						IAAS: "gcp",
						GCP: storage.GCP{
							ServiceAccountKey: serviceAccountKey,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "us-west1",
						},
						EnvID: "bbl-lake-time:stamp",
					})

					Expect(err).To(MatchError("failed to apply"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State.TFState).To(Equal("some-updated-tf-state"))
				})

				It("returns an error when the applier fails and we cannot retrieve the updated bbl state", func() {
					terraformManagerError.BBLStateCall.Returns.Error = errors.New("some-bbl-state-error")
					terraformManager.ApplyCall.Returns.Error = terraformManagerError

					err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
						IAAS: "gcp",
						GCP: storage.GCP{
							ServiceAccountKey: serviceAccountKey,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "us-west1",
						},
						EnvID: "bbl-lake-time:stamp",
					})

					Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nsome-bbl-state-error"))
					Expect(stateStore.SetCall.CallCount).To(Equal(2))
				})

				It("returns an error if applier fails with non terraform manager apply error", func() {
					terraformManager.ApplyCall.Returns.Error = errors.New("failed to apply")
					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					}, storage.State{})
					Expect(err).To(MatchError("failed to apply"))
				})

				It("returns an error when the terraform manager fails, we can retrieve the updated bbl state, and state fails to be set", func() {
					incomingState := storage.State{
						IAAS: "gcp",
						GCP: storage.GCP{
							ServiceAccountKey: serviceAccountKey,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "us-west1",
						},
						EnvID: "some-env-id",
					}
					updatedBBLState := incomingState
					updatedBBLState.TFState = "some-updated-tf-state"

					terraformManagerError.BBLStateCall.Returns.BBLState = updatedBBLState

					terraformManager.ApplyCall.Returns.Error = terraformManagerError

					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("state failed to be set")}}
					err := gcpUp.Execute(commands.GCPUpConfig{}, incomingState)

					Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nstate failed to be set"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State.TFState).To(Equal("some-updated-tf-state"))
				})
			})

			It("returns an error when the state fails to be set after applying terraform", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("state failed to be set")}}

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("state failed to be set"))
			})

			Context("bosh manager error handling", func() {
				Context("when bosh manager fails with bosh manager create error", func() {
					var (
						incomingState     storage.State
						expectedBOSHState map[string]interface{}
					)

					BeforeEach(func() {
						incomingState = storage.State{
							IAAS: "gcp",
							GCP: storage.GCP{
								ServiceAccountKey: serviceAccountKey,
								ProjectID:         "some-project-id",
								Zone:              "some-zone",
								Region:            "us-west1",
							},
							EnvID: "bbl-lake-time:stamp",
						}
						expectedBOSHState = map[string]interface{}{
							"partial": "bosh-state",
						}

						newState := incomingState
						newState.BOSH.State = expectedBOSHState

						expectedError := bosh.NewManagerCreateError(newState, errors.New("failed to create"))
						boshManager.CreateCall.Returns.Error = expectedError
					})

					It("returns the error and saves the state", func() {
						err := gcpUp.Execute(commands.GCPUpConfig{}, incomingState)
						Expect(err).To(MatchError("failed to create"))
						Expect(stateStore.SetCall.CallCount).To(Equal(4))
						Expect(stateStore.SetCall.Receives[3].State.BOSH.State).To(Equal(expectedBOSHState))
					})

					It("returns a compound error when it fails to save the state", func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("state failed to be set")}}

						err := gcpUp.Execute(commands.GCPUpConfig{
							ServiceAccountKey: serviceAccountKeyPath,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "us-west1",
						}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\nfailed to create,\nstate failed to be set"))
						Expect(stateStore.SetCall.CallCount).To(Equal(4))
						Expect(stateStore.SetCall.Receives[3].State.BOSH.State).To(Equal(expectedBOSHState))
					})
				})

				It("returns an error when bosh manager fails to create a bosh with a non bosh manager create error", func() {
					boshManager.CreateCall.Returns.Error = errors.New("failed to create")

					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKey: serviceAccountKeyPath,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					}, storage.State{})
					Expect(err).To(MatchError("failed to create"))
				})
			})

			It("returns an error when the state fails to be set after deploying bosh", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("state failed to be set")}}

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("state failed to be set"))
			})

			It("returns an error when the cloud config manager fails to update", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update")
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: serviceAccountKeyPath,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				}, storage.State{})
				Expect(err).To(MatchError("failed to update"))
			})
		})
	})
})
