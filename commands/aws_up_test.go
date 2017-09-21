package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSUp", func() {
	Describe("Execute", func() {
		var (
			command            commands.AWSUp
			boshManager        *fakes.BOSHManager
			terraformManager   *fakes.TerraformManager
			cloudConfigManager *fakes.CloudConfigManager
			stateStore         *fakes.StateStore
			envIDManager       *fakes.EnvIDManager
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}
			terraformManager.ApplyCall.Returns.BBLState = storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID:   "bbl-lake-time-stamp",
				TFState: "some-tf-state",
			}

			boshManager = &fakes.BOSHManager{}
			boshManager.CreateDirectorCall.Returns.State = storage.State{
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
			stateStore = &fakes.StateStore{}

			envIDManager = &fakes.EnvIDManager{}
			envIDManager.SyncCall.Returns.State = storage.State{
				EnvID: "bbl-lake-time-stamp",
			}

			command = commands.NewAWSUp(boshManager,
				cloudConfigManager, stateStore,
				envIDManager, terraformManager)
		})

		It("calls the env id manager and saves the env id", func() {
			err := command.Execute(commands.UpConfig{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 2))
			Expect(stateStore.SetCall.Receives[1].State.EnvID).To(Equal("bbl-lake-time-stamp"))
		})

		Context("when a name is passed in for env-id", func() {
			It("passes that name in for the env id manager to use", func() {
				err := command.Execute(commands.UpConfig{
					Name: "some-other-env-id",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
				Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-other-env-id"))
			})
		})

		It("creates infrastructure", func() {
			incomingState := storage.State{
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
			}

			err := command.Execute(commands.UpConfig{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(storage.State{
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
			}))

			Expect(stateStore.SetCall.CallCount).To(Equal(3))
			Expect(stateStore.SetCall.Receives[1].State).To(Equal(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID:   "bbl-lake-time-stamp",
				TFState: "some-tf-state",
			}))

			Expect(boshManager.CreateJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.CreateJumpboxCall.Receives.State).To(Equal(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID:   "bbl-lake-time-stamp",
				TFState: "some-tf-state",
				Jumpbox: storage.Jumpbox{},
			}))
		})

		Context("failure cases", func() {
			Context("when the terraform manager fails with terraformManagerError", func() {
				var (
					managerError *fakes.TerraformManagerError
				)

				BeforeEach(func() {
					managerError = &fakes.TerraformManagerError{}
					managerError.BBLStateCall.Returns.BBLState = storage.State{
						TFState: "some-partial-tf-state",
					}
					managerError.ErrorCall.Returns = "cannot apply"
					terraformManager.ApplyCall.Returns.Error = managerError
				})

				It("saves the bbl state and returns the error", func() {
					err := command.Execute(commands.UpConfig{}, storage.State{})
					Expect(err).To(MatchError("cannot apply"))

					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State).To(Equal(storage.State{
						TFState: "some-partial-tf-state",
					}))
				})

				Context("when the terraform manager error fails to return a bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.Error = errors.New("failed to retrieve bbl state")
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute(commands.UpConfig{}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\ncannot apply,\nfailed to retrieve bbl state"))
					})
				})

				Context("when we fail to set the bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.BBLState = storage.State{
							TFState: "some-partial-tf-state",
						}
						stateStore.SetCall.Returns = []fakes.SetCallReturn{
							{},
							{errors.New("failed to set bbl state")},
						}
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute(commands.UpConfig{}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\ncannot apply,\nfailed to set bbl state"))
					})
				})
			})

			Context("when the terraform manager fails with non terraformManagerError", func() {
				It("returns the error", func() {
					terraformManager.ApplyCall.Returns.Error = errors.New("cannot apply")

					err := command.Execute(commands.UpConfig{}, storage.State{})
					Expect(err).To(MatchError("cannot apply"))
				})
			})

			Context("when the state cannot be set", func() {
				It("returns the error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{
						{},
						{},
						{errors.New("failed to set the state")},
					}

					err := command.Execute(commands.UpConfig{}, storage.State{})
					Expect(err).To(MatchError("failed to set the state"))
				})
			})
		})

		Context("when the no-director flag is provided", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true
			})

			It("does not create a bosh or cloud config", func() {
				err := command.Execute(commands.UpConfig{
					NoDirector: true,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[1].State.NoDirector).To(BeTrue())
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
			})

			Context("when a bbl environment exists with no bosh director", func() {
				It("does not create a bosh director on subsequent runs", func() {
					err := command.Execute(commands.UpConfig{}, storage.State{
						NoDirector: true,
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
					Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
					Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
					Expect(stateStore.SetCall.CallCount).To(Equal(2))
				})
			})

			Context("when a bbl environment exists with a bosh director", func() {
				It("fast fails before creating any infrastructure", func() {
					err := command.Execute(commands.UpConfig{
						NoDirector: true,
					}, storage.State{
						BOSH: storage.BOSH{
							DirectorName: "some-director",
						},
					})

					Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
				})
			})
		})

		It("deploys bosh", func() {
			incomingState := storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
				},
				EnvID:   "bbl-lake-time-stamp",
				TFState: "some-tf-state",
			}

			err := command.Execute(commands.UpConfig{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))
			Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(incomingState))
		})

		Context("when ops file are passed in via --ops-file flag", func() {
			It("passes the ops file contents to the bosh manager", func() {
				opsFile, err := ioutil.TempFile("", "ops-file")
				Expect(err).NotTo(HaveOccurred())

				opsFilePath := opsFile.Name()
				opsFileContents := "some-ops-file-contents"
				err = ioutil.WriteFile(opsFilePath, []byte(opsFileContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = command.Execute(commands.UpConfig{
					OpsFile: opsFilePath,
				}, storage.State{
					EnvID: "bbl-lake-time-stamp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateDirectorCall.Receives.State.BOSH.UserOpsFile).To(Equal("some-ops-file-contents"))
			})
		})

		Describe("cloud config", func() {
			It("updates the bosh director with a cloud config provided an up-to-date state", func() {
				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(storage.State{
					EnvID: "bbl-lake-time-stamp",
					IAAS:  "aws",
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
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-aws-region",
					},
					TFState: "some-tf-state",
				}))
			})
		})

		Describe("state manipulation", func() {
			Context("iaas", func() {
				It("writes iaas aws to state", func() {
					err := command.Execute(commands.UpConfig{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State.IAAS).To(Equal("aws"))
				})
			})

			Context("aws credentials", func() {
				It("does not override the credentials when they're not passed in", func() {
					err := command.Execute(commands.UpConfig{}, storage.State{
						AWS: storage.AWS{
							AccessKeyID:     "aws-access-key-id",
							SecretAccessKey: "aws-secret-access-key",
							Region:          "aws-region",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(stateStore.SetCall.Receives[0].State.AWS).To(Equal(storage.AWS{
						AccessKeyID:     "aws-access-key-id",
						SecretAccessKey: "aws-secret-access-key",
						Region:          "aws-region",
					}))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the env id manager fails", func() {
				envIDManager.SyncCall.Returns.Error = errors.New("env id manager failed")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("env id manager failed"))

			})

			It("returns an error when saving the state fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{
					{
						Error: errors.New("saving the state failed"),
					},
				}
				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("saving the state failed"))
			})

			It("returns an error when the cloud config cannot be uploaded", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update")
				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to update"))
			})

			It("returns an error when the terraform manager cannot get terraform outputs", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("cannot parse terraform output")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("cannot parse terraform output"))
			})

			It("returns an error when the ops file cannot be read", func() {
				err := command.Execute(commands.UpConfig{
					OpsFile: "some/fake/path",
				}, storage.State{})
				Expect(err).To(MatchError("error reading ops-file contents: open some/fake/path: no such file or directory"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshManager.CreateDirectorCall.Returns.Error = errors.New("cannot deploy bosh")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("cannot deploy bosh"))
			})

			It("returns an error when state store fails to set the state before retrieving availability zones", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before updating the cloud config", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("failed to set state")}}

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			Context("when the bosh manager fails with BOSHManagerCreate error", func() {
				var (
					incomingState     storage.State
					expectedBOSHState map[string]interface{}
				)

				BeforeEach(func() {
					incomingState = storage.State{
						IAAS: "aws",
						AWS: storage.AWS{
							Region:          "some-aws-region",
							SecretAccessKey: "some-secret-access-key",
							AccessKeyID:     "some-access-key-id",
						},
						EnvID: "bbl-lake-time:stamp",
					}
					expectedBOSHState = map[string]interface{}{
						"partial": "bosh-state",
					}

					newState := incomingState
					newState.BOSH.State = expectedBOSHState
					expectedError := bosh.NewManagerCreateError(newState, errors.New("failed to create"))
					boshManager.CreateDirectorCall.Returns.Error = expectedError
				})

				It("returns the error and saves the state", func() {
					err := command.Execute(commands.UpConfig{}, incomingState)
					Expect(err).To(MatchError("failed to create"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State.BOSH.State).To(Equal(expectedBOSHState))
				})

				It("returns a compound error when it fails to save the state", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("state failed to be set")}}
					err := command.Execute(commands.UpConfig{}, incomingState)
					Expect(err).To(MatchError("the following errors occurred:\nfailed to create,\nstate failed to be set"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State.BOSH.State).To(Equal(expectedBOSHState))
				})
			})
		})
	})
})
