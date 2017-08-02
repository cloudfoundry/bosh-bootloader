package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/cloudfoundry/multierror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPCreateLBs", func() {
	var (
		terraformManager          *fakes.TerraformManager
		cloudConfigManager        *fakes.CloudConfigManager
		stateStore                *fakes.StateStore
		logger                    *fakes.Logger
		terraformExecutorError    *fakes.TerraformExecutorError
		availabilityZoneRetriever *fakes.Zones

		bblState    storage.State
		command     commands.GCPCreateLBs
		certPath    string
		keyPath     string
		certificate string
		key         string
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		stateStore = &fakes.StateStore{}
		logger = &fakes.Logger{}
		terraformExecutorError = &fakes.TerraformExecutorError{}
		availabilityZoneRetriever = &fakes.Zones{}

		command = commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, logger, availabilityZoneRetriever)

		tempCertFile, err := ioutil.TempFile("", "cert")
		Expect(err).NotTo(HaveOccurred())

		certificate = "some-cert"
		certPath = tempCertFile.Name()
		err = ioutil.WriteFile(certPath, []byte(certificate), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		tempKeyFile, err := ioutil.TempFile("", "key")
		Expect(err).NotTo(HaveOccurred())

		key = "some-key"
		keyPath = tempKeyFile.Name()
		err = ioutil.WriteFile(keyPath, []byte(key), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		bblState = storage.State{
			IAAS: "gcp",
			GCP: storage.GCP{
				Region: "some-region",
			},
			BOSH: storage.BOSH{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				DirectorAddress:  "some-director-address",
			},
			TFState: "some-tfstate",
		}
	})

	AfterEach(func() {
		commands.ResetMarshal()
	})

	Describe("Execute", func() {
		Context("when lb type is cf", func() {
			BeforeEach(func() {
				availabilityZoneRetriever.GetCall.Returns.Zones = []string{"z1", "z2", "z3"}
			})
			It("calls terraform manager apply", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
					Domain:   "some-domain",
				}, bblState)
				Expect(err).NotTo(HaveOccurred())

				By("getting the AZs", func() {
					Expect(availabilityZoneRetriever.GetCall.Receives.Region).To(Equal("some-region"))
				})

				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						Zones:  []string{"z1", "z2", "z3"},
						Region: "some-region",
					},
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						DirectorAddress:  "some-director-address",
					},
					TFState: "some-tfstate",
					LB: storage.LB{
						Type:   "cf",
						Cert:   certificate,
						Key:    key,
						Domain: "some-domain",
					},
				}))
			})
		})

		Context("when lb type is concourse", func() {
			It("calls terraform manager apply", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS:    "gcp",
					TFState: "some-tfstate",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
					TFState: "some-tfstate",
				}))
			})
		})

		It("saves the updated tfstate", func() {
			terraformManager.ApplyCall.Returns.BBLState = storage.State{
				IAAS: "gcp",
				BOSH: storage.BOSH{
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
					DirectorAddress:  "some-director-address",
				},
				LB: storage.LB{
					Type: "concourse",
				},
				TFState: "some-new-tfstate",
			}

			err := command.Execute(commands.GCPCreateLBsConfig{
				LBType: "concourse",
			}, storage.State{
				IAAS:    "gcp",
				TFState: "some-old-tfstate",
				BOSH: storage.BOSH{
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
					DirectorAddress:  "some-director-address",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
				IAAS: "gcp",
				BOSH: storage.BOSH{
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
					DirectorAddress:  "some-director-address",
				},
				LB: storage.LB{
					Type: "concourse",
				},
				TFState: "some-new-tfstate",
			}))
		})

		It("uploads a new cloud-config to the bosh director", func() {
			terraformManager.ApplyCall.Returns.BBLState = bblState

			err := command.Execute(commands.GCPCreateLBsConfig{
				LBType: "concourse",
			}, bblState)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(bblState))
		})

		It("no-ops if SkipIfExists is supplied and the LBType does not change", func() {
			bblState.LB.Type = "concourse"
			err := command.Execute(commands.GCPCreateLBsConfig{
				LBType:       "concourse",
				SkipIfExists: true,
			}, bblState)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Messages).To(ContainElement(`lb type "concourse" exists, skipping...`))
			Expect(terraformManager.ApplyCall.CallCount).To(Equal(0))
			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
		})

		Context("when there is no BOSH director", func() {
			It("creates the LBs", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS:       "gcp",
					EnvID:      "some-env-id",
					TFState:    "some-prev-tf-state",
					NoDirector: true,
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not call the CloudConfigManager", func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS:       "gcp",
					EnvID:      "some-env-id",
					TFState:    "some-prev-tf-state",
					NoDirector: true,
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when creating a cf lb and provided cert and key files are empty", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(certPath, []byte{}, os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(keyPath, []byte{}, os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns a helpful error message", func() {
					expectedErrors := multierror.NewMultiError("create-lbs")
					expectedErrors.Add(errors.New("provided cert file is empty"))
					expectedErrors.Add(errors.New("provided key file is empty"))

					err := command.Execute(commands.GCPCreateLBsConfig{
						LBType:   "cf",
						CertPath: certPath,
						KeyPath:  keyPath,
						Domain:   "some-domain",
					}, storage.State{
						IAAS: "gcp",
					})
					Expect(err).To(Equal(expectedErrors))
				})
			})

			It("returns an error if terraform manager version validator fails", func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("cannot validate version")

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS: "gcp",
				})

				Expect(err).To(MatchError("cannot validate version"))
			})

			It("returns a helpful error when no lb type is provided", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("--type is a required flag"))
			})

			It("returns an error when the lb type is not concourse or cf", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "some-fake-lb",
				}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError(`"some-fake-lb" is not a valid lb type, valid lb types are: concourse, cf`))
			})

			Context("tf state is empty", func() {
				It("returns a BBLNotFound error", func() {
					err := command.Execute(commands.GCPCreateLBsConfig{
						LBType: "concourse",
					}, storage.State{IAAS: "gcp"})
					Expect(err).To(MatchError(commands.BBLNotFound))
				})
			})

			Context("when lb type is cf", func() {
				It("returns an error when cert and key are not provided", func() {
					expectedErrors := multierror.NewMultiError("create-lbs")
					expectedErrors.Add(errors.New("--cert is required"))
					expectedErrors.Add(errors.New("--key is required"))

					err := command.Execute(commands.GCPCreateLBsConfig{
						LBType: "cf",
					}, storage.State{IAAS: "gcp", TFState: "some-tf-state"})
					Expect(err).To(MatchError(expectedErrors))
				})
			})

			It("returns an error when the iaas type is not gcp", func() {
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "aws"})
				Expect(err).To(MatchError("iaas type must be gcp"))
			})

			It("returns an error when the availability zone retriever fails to get zones", func() {
				availabilityZoneRetriever.GetCall.Returns.Error = errors.New("failed to get zones")
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp", TFState: "some-tf-state"})
				Expect(err).To(MatchError("failed to get zones"))
			})

			It("returns an error if the command fails to read the certificate", func() {
				expectedErrors := multierror.NewMultiError("create-lbs")
				expectedErrors.Add(errors.New("open some/fake/path: no such file or directory"))

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: "some/fake/path",
					KeyPath:  keyPath,
				}, storage.State{IAAS: "gcp", TFState: "some-tf-state"})
				Expect(err).To(MatchError(expectedErrors))
			})

			It("returns an error if the command fails to read the key", func() {
				expectedErrors := multierror.NewMultiError("create-lbs")
				expectedErrors.Add(errors.New("open some/fake/path: no such file or directory"))

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  "some/fake/path",
				}, storage.State{IAAS: "gcp", TFState: "some-tf-state"})
				Expect(err).To(MatchError(expectedErrors))
			})

			It("saves the tf state even if the applier fails", func() {
				terraformExecutorError.TFStateCall.Returns.TFState = "some-updated-tf-state"
				terraformExecutorError.ErrorCall.Returns = "failed to apply"
				expectedError := terraform.NewManagerError(storage.State{
					TFState: "some-tf-state",
				}, terraformExecutorError)
				terraformManager.ApplyCall.Returns.Error = expectedError

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{
					IAAS:    "gcp",
					EnvID:   "some-env-id",
					TFState: "some-prev-tf-state",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})

				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State.TFState).To(Equal("some-updated-tf-state"))
			})

			It("returns an error if terraform manager apply fails with non terraform manager apply error", func() {
				terraformManager.ApplyCall.Returns.Error = errors.New("failed to apply")
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: certPath,
					KeyPath:  keyPath,
				}, storage.State{
					IAAS:    "gcp",
					TFState: "some-tf-state",
				})
				Expect(err).To(MatchError("failed to apply"))
				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			It("returns an error when both the applier fails and terraformManagerError.BBLState fails", func() {
				terraformExecutorError.TFStateCall.Returns.Error = errors.New("failed to get tf state")
				terraformExecutorError.ErrorCall.Returns = "failed to apply"
				expectedError := terraform.NewManagerError(bblState, terraformExecutorError)
				terraformManager.ApplyCall.Returns.Error = expectedError

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, bblState)

				Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nfailed to get tf state"))
				Expect(stateStore.SetCall.CallCount).To(Equal(0))
			})

			It("returns an error when both the applier fails and state fails to be set", func() {
				terraformExecutorError.TFStateCall.Returns.TFState = "some-updated-tf-state"
				terraformExecutorError.ErrorCall.Returns = "failed to apply"
				expectedError := terraform.NewManagerError(storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
					TFState: "some-tf-state",
				}, terraformExecutorError)
				terraformManager.ApplyCall.Returns.Error = expectedError

				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("state failed to be set")}}
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp", TFState: "some-tf-state"})

				Expect(err).To(MatchError("the following errors occurred:\nfailed to apply,\nstate failed to be set"))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
					TFState: "some-updated-tf-state",
				}))
			})

			It("returns an error when the state store fails to save the state after applying terraform", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{fakes.SetCallReturn{Error: errors.New("failed to save state")}}
				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp", TFState: "some-tf-state"})

				Expect(err).To(MatchError("failed to save state"))
			})

			It("returns an error when the cloud config fails to be updated", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update cloud config")

				err := command.Execute(commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}, storage.State{IAAS: "gcp", TFState: "some-tf-state"})
				Expect(err).To(MatchError("failed to update cloud config"))
			})
		})
	})
})
