package commands_test

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	compute "google.golang.org/api/compute/v1"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

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
		gcpUp              commands.GCPUp
		stateStore         *fakes.StateStore
		keyPairUpdater     *fakes.GCPKeyPairUpdater
		gcpClientProvider  *fakes.GCPClientProvider
		gcpClient          *fakes.GCPClient
		terraformExecutor  *fakes.TerraformExecutor
		boshManager        *fakes.BOSHManager
		cloudConfigManager *fakes.CloudConfigManager
		envIDManager       *fakes.EnvIDManager
		logger             *fakes.Logger
		zones              *fakes.Zones

		serviceAccountKeyPath     string
		serviceAccountKey         string
		expectedTerraformTemplate string
	)

	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		keyPairUpdater = &fakes.GCPKeyPairUpdater{}
		gcpClientProvider = &fakes.GCPClientProvider{}
		gcpClient = &fakes.GCPClient{}
		gcpClientProvider.ClientCall.Returns.Client = gcpClient
		gcpClient.GetNetworksCall.Returns.NetworkList = &compute.NetworkList{}
		terraformExecutor = &fakes.TerraformExecutor{}
		terraformExecutor.VersionCall.Returns.Version = "0.8.7"
		zones = &fakes.Zones{}
		envIDManager = &fakes.EnvIDManager{}
		terraformExecutor.ApplyCall.Returns.TFState = "some-tf-state"
		envIDManager.SyncCall.Returns.EnvID = "some-env-id"

		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}
		boshManager.CreateCall.Returns.State = storage.State{
			BOSH: storage.BOSH{
				DirectorName:           "bosh-bbl-lake-time:stamp",
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
			},
		}
		cloudConfigManager = &fakes.CloudConfigManager{}
		gcpUp = commands.NewGCPUp(stateStore, keyPairUpdater, gcpClientProvider, terraformExecutor, boshManager,
			logger, zones, envIDManager, cloudConfigManager)

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		serviceAccountKey = `{"real": "json"}`
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		body, err := ioutil.ReadFile("fixtures/terraform_template_no_lb.tf")
		Expect(err).NotTo(HaveOccurred())

		expectedTerraformTemplate = string(body)
	})

	AfterEach(func() {
		commands.ResetMarshal()
	})

	Describe("Execute", func() {
		It("saves gcp details to the state", func() {
			keyPairUpdater.UpdateCall.Returns.KeyPair = storage.KeyPair{
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.Receives.State.IAAS).To(Equal("gcp"))
			Expect(stateStore.SetCall.Receives.State.GCP).To(Equal(storage.GCP{
				ServiceAccountKey: serviceAccountKey,
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "us-west1",
			}))
			Expect(stateStore.SetCall.Receives.State.KeyPair).To(Equal(storage.KeyPair{
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}))
			Expect(gcpClientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(gcpClientProvider.SetConfigCall.Receives.ServiceAccountKey).To(Equal(`{"real": "json"}`))
			Expect(gcpClientProvider.SetConfigCall.Receives.ProjectID).To(Equal("some-project-id"))
			Expect(gcpClientProvider.SetConfigCall.Receives.Zone).To(Equal("some-zone"))
		})

		It("calls env id manager and saves the resulting state", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives.State.EnvID).To(Equal("some-env-id"))
		})

		Context("when a name is passed in for env-id", func() {
			It("passes that name in for the env id manager to use", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "us-west1",
					Name:                  "some-other-env-id",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
				Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-other-env-id"))
			})
		})

		It("uploads the ssh keys", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(1))
		})

		Context("when ops file are passed in via --ops-file flag", func() {
			It("passes the ops file contents to the bosh manager", func() {
				opsFile, err := ioutil.TempFile("", "ops-file")
				Expect(err).NotTo(HaveOccurred())

				opsFilePath := opsFile.Name()
				opsFileContents := "some-ops-file-contents"
				err = ioutil.WriteFile(opsFilePath, []byte(opsFileContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "us-west1",
					OpsFilePath:           opsFilePath,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateCall.Receives.OpsFile).To(Equal([]byte("some-ops-file-contents")))
			})
		})

		It("creates gcp resources via terraform", func() {
			gcpUpConfig := commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}

			err := gcpUp.Execute(gcpUpConfig, storage.State{
				EnvID: "some-env-id",
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformExecutor.ApplyCall.Receives.Credentials).To(Equal(serviceAccountKey))
			Expect(terraformExecutor.ApplyCall.Receives.EnvID).To(Equal("some-env-id"))
			Expect(terraformExecutor.ApplyCall.Receives.ProjectID).To(Equal("some-project-id"))
			Expect(terraformExecutor.ApplyCall.Receives.Zone).To(Equal("some-zone"))
			Expect(terraformExecutor.ApplyCall.Receives.Region).To(Equal("us-west1"))
			Expect(terraformExecutor.ApplyCall.Receives.Template).To(Equal(expectedTerraformTemplate))
			Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-tf-state"))
		})

		Context("when the no-director flag is provided", func() {
			It("does not create a bosh or cloud config", func() {
				gcpUpConfig := commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "us-west1",
					NoDirector:            true,
				}

				err := gcpUp.Execute(gcpUpConfig, storage.State{})

				Expect(err).NotTo(HaveOccurred())

				Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateCall.CallCount).To(Equal(0))
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
				Expect(stateStore.SetCall.CallCount).To(Equal(3))
				Expect(stateStore.SetCall.Receives.State.NoDirector).To(Equal(true))
			})

			Context("when a bbl environment exists with a bosh director", func() {
				It("fast fails before creating any infrastructure", func() {
					gcpUpConfig := commands.GCPUpConfig{
						ServiceAccountKeyPath: serviceAccountKeyPath,
						ProjectID:             "some-project-id",
						Zone:                  "some-zone",
						Region:                "us-west1",
						NoDirector:            true,
					}

					err := gcpUp.Execute(gcpUpConfig, storage.State{
						BOSH: storage.BOSH{
							DirectorName: "some-director",
						},
					})

					Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
				})
			})

			Context("when re-bbling up an environment with no director", func() {
				It("is does not create a bosh director", func() {
					gcpUpConfig := commands.GCPUpConfig{
						ServiceAccountKeyPath: serviceAccountKeyPath,
						ProjectID:             "some-project-id",
						Zone:                  "some-zone",
						Region:                "us-west1",
					}

					err := gcpUp.Execute(gcpUpConfig, storage.State{
						NoDirector: true,
					})

					Expect(err).NotTo(HaveOccurred())

					Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
					Expect(boshManager.CreateCall.CallCount).To(Equal(0))
					Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives.State.NoDirector).To(Equal(true))
				})
			})
		})

		Describe("bosh", func() {
			It("creates a bosh", func() {
				envIDManager.SyncCall.Returns.EnvID = "bbl-lake-time:stamp"

				err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
					IAAS: "gcp",
					KeyPair: storage.KeyPair{
						PublicKey:  "some-public-key",
						PrivateKey: "some-private-key",
					},
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					},
					EnvID: "bbl-lake-time:stamp",
					BOSH: storage.BOSH{
						State: map[string]interface{}{
							"new-key": "new-value",
						},
						Variables: variablesYAML,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
					KeyPair: storage.KeyPair{
						PublicKey:  "some-public-key",
						PrivateKey: "some-private-key",
					},
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "us-west1",
					},
					EnvID: "bbl-lake-time:stamp",
					BOSH: storage.BOSH{
						State: map[string]interface{}{
							"new-key": "new-value",
						},
						Variables: variablesYAML,
					},
					TFState: "some-tf-state",
				}))
			})

			Describe("state manipulation", func() {
				Context("when the state file does not exist", func() {
					BeforeEach(func() {
						err := gcpUp.Execute(commands.GCPUpConfig{
							ServiceAccountKeyPath: serviceAccountKeyPath,
							ProjectID:             "some-project-id",
							Zone:                  "some-zone",
							Region:                "us-west1",
						}, storage.State{
							EnvID: "bbl-lake-time:stamp",
						})
						Expect(err).NotTo(HaveOccurred())
					})

					It("saves the bosh manager create state and variables", func() {
						Expect(stateStore.SetCall.Receives.State.BOSH).To(Equal(storage.BOSH{
							DirectorName:           "bosh-bbl-lake-time:stamp",
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
						}))
					})
				})

				Context("when the state file exists", func() {
					BeforeEach(func() {
						err := gcpUp.Execute(commands.GCPUpConfig{
							ServiceAccountKeyPath: serviceAccountKeyPath,
							ProjectID:             "some-project-id",
							Zone:                  "some-zone",
							Region:                "us-west1",
						}, storage.State{
							EnvID: "bbl-lake-time:stamp",
							BOSH: storage.BOSH{
								State: map[string]interface{}{
									"old-key": "old-value",
								},
							},
						})
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the bosh state", func() {
						Expect(stateStore.SetCall.Receives.State.BOSH.State).To(Equal(map[string]interface{}{
							"new-key": "new-value",
						}))
					})
				})
			})

			Describe("bosh-related failure cases", func() {
				It("returns an error when the ops file cannot be read", func() {
					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKeyPath: serviceAccountKeyPath,
						OpsFilePath:           "some/fake/path",
					}, storage.State{})
					Expect(err).To(MatchError("error reading ops-file contents: open some/fake/path: no such file or directory"))
				})

				It("returns an error when bosh manager fails to create a bosh", func() {
					boshManager.CreateCall.Returns.Error = errors.New("failed to create")

					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKeyPath: serviceAccountKeyPath,
						ProjectID:             "some-project-id",
						Zone:                  "some-zone",
						Region:                "us-west1",
					}, storage.State{})
					Expect(err).To(MatchError("failed to create"))
				})

				It("returns an error when the state fails to be set after deploying bosh", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("state failed to be set")}}

					err := gcpUp.Execute(commands.GCPUpConfig{
						ServiceAccountKeyPath: serviceAccountKeyPath,
						ProjectID:             "some-project-id",
						Zone:                  "some-zone",
						Region:                "us-west1",
					}, storage.State{})
					Expect(err).To(MatchError("state failed to be set"))
				})
			})
		})
	})

	Describe("cloud config", func() {
		It("updates the cloud config", func() {
			zones.GetCall.Returns.Zones = []string{"zone-1", "zone-2", "zone-3"}
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "some-region",
			}, storage.State{
				EnvID: "bbl-lake-time:stamp",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: serviceAccountKey,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				KeyPair: storage.KeyPair{
					Name:       "",
					PrivateKey: "",
					PublicKey:  "",
				},
				BOSH: storage.BOSH{
					DirectorName:           "bosh-bbl-lake-time:stamp",
					DirectorUsername:       "admin",
					DirectorPassword:       "some-admin-password",
					DirectorAddress:        "some-director-address",
					DirectorSSLCA:          "some-ca",
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					Variables:              variablesYAML,
					State: map[string]interface{}{
						"new-key": "new-value",
					},
					Manifest: "some-bosh-manifest",
				},
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
			}))
		})
	})

	Context("when state contains gcp details", func() {
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

		It("does not create a new ssh key", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: serviceAccountKey,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				},
				KeyPair: storage.KeyPair{
					Name: "some-key-name",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(0))
		})

		It("calls terraform executor with previous tf state", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: serviceAccountKey,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				},
				TFState: "some-tf-state",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformExecutor.ApplyCall.Receives.TFState).To(Equal("some-tf-state"))
		})

		It("does not require details from up config", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: serviceAccountKey,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "us-west1",
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("up config contains subset of the details",
			func(upConfig func() commands.GCPUpConfig, expectedErr string) {
				err := gcpUp.Execute(upConfig(), storage.State{
					IAAS: "gcp",
					GCP:  storage.GCP{},
				})
				Expect(err).To(MatchError(expectedErr))
			},
			Entry("returns an error when the service account key is not provided", func() commands.GCPUpConfig {
				return commands.GCPUpConfig{
					ProjectID: "new-project-id",
					Zone:      "new-zone",
					Region:    "new-region",
				}
			}, "GCP service account key must be provided"),
			Entry("returns an error when the project ID is not provided", func() commands.GCPUpConfig {
				return commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					Zone:   "new-zone",
					Region: "new-region",
				}
			}, "GCP project ID must be provided"),
			Entry("returns an error when the zone is not provided", func() commands.GCPUpConfig {
				return commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "new-project-id",
					Region:                "new-region",
				}
			}, "GCP zone must be provided"),
			Entry("returns an error when the region is not provided", func() commands.GCPUpConfig {
				return commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "new-project-id",
					Zone:                  "new-zone",
				}
			}, "GCP region must be provided"),
		)
	})

	Context("when lb type exists in the state", func() {
		Context("when a cf lb exists in the state", func() {
			BeforeEach(func() {
				zones.GetCall.Returns.Zones = []string{"some-zone", "some-other-zone"}
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "some-region",
				}, storage.State{
					LB: storage.LB{
						Type:   "cf",
						Cert:   "some-cert",
						Key:    "some-key",
						Domain: "some-domain",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("applies the correct cf template and args for cf lb type", func() {
				Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformExecutor.ApplyCall.Receives.Template).To(Equal(strings.Join([]string{expectedCFTemplate, dnsTemplate}, "\n")))
				Expect(terraformExecutor.ApplyCall.Receives.Cert).To(Equal("some-cert"))
				Expect(terraformExecutor.ApplyCall.Receives.Key).To(Equal("some-key"))
				Expect(terraformExecutor.ApplyCall.Receives.Domain).To(Equal("some-domain"))
			})
		})

		Context("when a concourse lb exists in the state", func() {
			BeforeEach(func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "some-region",
				}, storage.State{
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("applies the correct concourse template and args for concourse lb type", func() {
				Expect(terraformExecutor.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformExecutor.ApplyCall.Receives.Template).To(Equal(expectedConcourseTemplate))
				Expect(terraformExecutor.ApplyCall.Receives.Cert).To(Equal(""))
				Expect(terraformExecutor.ApplyCall.Receives.Key).To(Equal(""))
				Expect(terraformExecutor.ApplyCall.Receives.Domain).To(Equal(""))
			})
		})
	})

	Context("failure cases", func() {
		It("fast fails if a gcp environment with the same name already exists", func() {
			envIDManager.SyncCall.Returns.Error = errors.New("environment already exists")
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})

			Expect(err).To(MatchError("environment already exists"))
		})

		It("fast fails if the terraform installed is less than v0.8.5", func() {
			terraformExecutor.VersionCall.Returns.Version = "0.8.4"

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})

			Expect(err).To(MatchError("Terraform version must be at least v0.8.5"))
		})

		It("fast fails if the terraform executor fails to get the version", func() {
			terraformExecutor.VersionCall.Returns.Error = errors.New("cannot get version")

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})

			Expect(err).To(MatchError("cannot get version"))
		})

		It("fast fails when the major version cannot be converted to an int", func() {
			terraformExecutor.VersionCall.Returns.Version = "lol.5.2"

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})

			Expect(err.Error()).To(ContainSubstring("invalid syntax"))
		})

		It("fast fails when the minor version cannot be converted to an int", func() {
			terraformExecutor.VersionCall.Returns.Version = "0.lol.2"

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})

			Expect(err.Error()).To(ContainSubstring("invalid syntax"))
		})

		It("fast fails when the patch version cannot be converted to an int", func() {
			terraformExecutor.VersionCall.Returns.Version = "0.5.lol"

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})

			Expect(err.Error()).To(ContainSubstring("invalid syntax"))
		})

		It("returns an error if applier fails with non terraform apply error", func() {
			terraformExecutor.ApplyCall.Returns.Error = errors.New("failed to apply")
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{
				IAAS: "gcp",
				Stack: storage.Stack{
					LBType: "concourse",
				},
			})
			Expect(err).To(MatchError("failed to apply"))
			Expect(stateStore.SetCall.CallCount).To(Equal(2))
		})

		Describe("calling up with different gcp flags then the state", func() {
			It("returns an error when the --gcp-region is different", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "some-other-region",
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
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-other-zone",
					Region:                "us-west1",
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
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-other-project-id",
					Zone:                  "some-zone",
					Region:                "us-west1",
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

		It("returns an error when state store fails", func() {
			stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("set call failed")}}
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "p",
				Zone:                  "z",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("set call failed"))
		})

		It("should not store the state if the provided flags are not valid", func() {
			err := gcpUp.Execute(
				commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
				}, storage.State{})
			Expect(err).To(MatchError("GCP project ID must be provided"))
			Expect(stateStore.SetCall.CallCount).To(Equal(0))
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
					ServiceAccountKeyPath: serviceAccountKeyPath,
					Zone:   "z",
					Region: "us-west1",
				}
			}, "GCP project ID must be provided"),
			Entry("returns an error when zone is missing", func() commands.GCPUpConfig {
				return commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "p",
					Region:                "us-west1",
				}
			}, "GCP zone must be provided"),
			Entry("returns an error when region is missing", func() commands.GCPUpConfig {
				return commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "p",
					Zone:                  "z",
				}
			}, "GCP region must be provided"),
		)

		It("returns an error when the service account key file does not exist", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: "/some/non/existent/file",
				ProjectID:             "p",
				Zone:                  "z",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("error reading service account key: open /some/non/existent/file: no such file or directory"))
		})

		It("returns an error when the service account key file does not contain valid json", func() {
			tempFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			invalidServiceAccountKeyPath := tempFile.Name()
			err = ioutil.WriteFile(invalidServiceAccountKeyPath, []byte(`%%%not-valid-json%%%`), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: invalidServiceAccountKeyPath,
				ProjectID:             "p",
				Zone:                  "z",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("error parsing service account key: invalid character '%' looking for beginning of value"))
		})

		It("returns an error when the keypair could not be updated", func() {
			keyPairUpdater.UpdateCall.Returns.Error = errors.New("keypair update failed")

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("keypair update failed"))
		})

		It("returns an error when setting config fails", func() {
			gcpClientProvider.SetConfigCall.Returns.Error = errors.New("setting config failed")

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("setting config failed"))
		})

		It("saves the keypair when the terraform fails", func() {
			terraformExecutor.ApplyCall.Returns.Error = errors.New("terraform executor failed")
			keyPairUpdater.UpdateCall.Returns.KeyPair = storage.KeyPair{
				Name: "some-key-pair",
			}

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("terraform executor failed"))

			Expect(stateStore.SetCall.Receives.State.KeyPair.IsEmpty()).To(BeFalse())
		})

		It("returns an error when terraform executor fails", func() {
			terraformExecutor.ApplyCall.Returns.Error = errors.New("terraform executor failed")

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("terraform executor failed"))
		})

		It("saves the tf state when the applier fails", func() {
			expectedError := terraform.NewTerraformApplyError("some-tf-state", errors.New("failed to apply"))
			terraformExecutor.ApplyCall.Returns.Error = expectedError

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
			Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-tf-state"))
		})

		Context("when the bosh manager fails with BOSHManagerCreate error", func() {
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
				Expect(stateStore.SetCall.Receives.State.BOSH.State).To(Equal(expectedBOSHState))
			})

			It("returns a compound error when it fails to save the state", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("state failed to be set")}}
				err := gcpUp.Execute(commands.GCPUpConfig{}, incomingState)
				Expect(err).To(MatchError("the following errors occurred:\nfailed to create,\nstate failed to be set"))
				Expect(stateStore.SetCall.CallCount).To(Equal(4))
				Expect(stateStore.SetCall.Receives.State.BOSH.State).To(Equal(expectedBOSHState))
			})
		})

		It("returns an error when the state fails to be set after updating keypair", func() {
			stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("state failed to be set")}}

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("state failed to be set"))
		})

		It("returns an error when both the applier fails and state fails to be set", func() {
			expectedError := terraform.NewTerraformApplyError("some-tf-state", errors.New("failed to apply"))
			terraformExecutor.ApplyCall.Returns.Error = expectedError
			terraformExecutor.ApplyCall.Returns.TFState = "some-tf-state"

			stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("state failed to be set")}}
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

			Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nstate failed to be set"))
			Expect(stateStore.SetCall.CallCount).To(Equal(3))
			Expect(stateStore.SetCall.Receives.State.TFState).To(Equal("some-tf-state"))
		})

		It("returns an error when the state fails to be set after applying terraform", func() {
			stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("state failed to be set")}}

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("state failed to be set"))
		})

		It("returns an error when the cloud config manager fails to update", func() {
			cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update")
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "us-west1",
			}, storage.State{})
			Expect(err).To(MatchError("failed to update"))
		})
	})
})
