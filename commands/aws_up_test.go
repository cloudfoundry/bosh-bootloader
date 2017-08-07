package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSUp", func() {
	Describe("Execute", func() {
		var (
			command                    commands.AWSUp
			boshManager                *fakes.BOSHManager
			terraformManager           *fakes.TerraformManager
			keyPairManager             *fakes.KeyPairManager
			credentialValidator        *fakes.CredentialValidator
			cloudConfigManager         *fakes.CloudConfigManager
			brokenEnvironmentValidator *fakes.BrokenEnvironmentValidator
			stateStore                 *fakes.StateStore
			awsClientProvider          *fakes.AWSClientProvider
			envIDManager               *fakes.EnvIDManager
		)

		BeforeEach(func() {
			keyPairManager = &fakes.KeyPairManager{}
			keyPairManager.SyncCall.Returns.KeyPair = storage.KeyPair{
				Name:       "keypair-bbl-lake-time-stamp",
				PublicKey:  "some-public-key",
				PrivateKey: "some-private-key",
			}

			terraformManager = &fakes.TerraformManager{}
			terraformManager.ApplyCall.Returns.BBLState = storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
				KeyPair: storage.KeyPair{
					Name:       "keypair-bbl-lake-time-stamp",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
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

			credentialValidator = &fakes.CredentialValidator{}

			stateStore = &fakes.StateStore{}
			awsClientProvider = &fakes.AWSClientProvider{}

			envIDManager = &fakes.EnvIDManager{}
			envIDManager.SyncCall.Returns.State = storage.State{
				EnvID: "bbl-lake-time-stamp",
			}

			brokenEnvironmentValidator = &fakes.BrokenEnvironmentValidator{}

			command = commands.NewAWSUp(
				credentialValidator, keyPairManager, boshManager,
				cloudConfigManager, stateStore, awsClientProvider,
				envIDManager, terraformManager, brokenEnvironmentValidator,
			)
		})

		It("returns an error when aws credential validator fails", func() {
			credentialValidator.ValidateCall.Returns.Error = errors.New("failed to validate aws credentials")
			err := command.Execute(commands.AWSUpConfig{}, storage.State{})
			Expect(err).To(MatchError("failed to validate aws credentials"))
		})

		It("retrieves a client with the provided credentials", func() {
			err := command.Execute(commands.AWSUpConfig{
				AccessKeyID:     "new-aws-access-key-id",
				SecretAccessKey: "new-aws-secret-access-key",
				Region:          "new-aws-region",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(awsClientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(awsClientProvider.SetConfigCall.Receives.Config).To(Equal(aws.Config{
				Region:          "new-aws-region",
				SecretAccessKey: "new-aws-secret-access-key",
				AccessKeyID:     "new-aws-access-key-id",
			}))
			Expect(credentialValidator.ValidateCall.CallCount).To(Equal(0))
		})

		It("calls the env id manager and saves the env id", func() {
			err := command.Execute(commands.AWSUpConfig{
				AccessKeyID:     "new-aws-access-key-id",
				SecretAccessKey: "new-aws-secret-access-key",
				Region:          "new-aws-region",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 2))
			Expect(stateStore.SetCall.Receives[1].State.EnvID).To(Equal("bbl-lake-time-stamp"))
		})

		Context("when a name is passed in for env-id", func() {
			It("passes that name in for the env id manager to use", func() {
				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "new-aws-access-key-id",
					SecretAccessKey: "new-aws-secret-access-key",
					Region:          "new-aws-region",
					Name:            "some-other-env-id",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
				Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-other-env-id"))
			})
		})

		It("syncs the keypair", func() {
			err := command.Execute(commands.AWSUpConfig{}, storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(awsClientProvider.SetConfigCall.CallCount).To(Equal(0))
			Expect(credentialValidator.ValidateCall.CallCount).To(Equal(1))

			Expect(keyPairManager.SyncCall.Receives.State).To(Equal(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
			}))

			Expect(stateStore.SetCall.CallCount).To(Equal(4))
			actualState := stateStore.SetCall.Receives[3].State
			Expect(actualState.KeyPair).To(Equal(storage.KeyPair{
				Name:       "keypair-bbl-lake-time-stamp",
				PublicKey:  "some-public-key",
				PrivateKey: "some-private-key",
			}))
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

			err := command.Execute(commands.AWSUpConfig{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
				KeyPair: storage.KeyPair{
					Name:       "keypair-bbl-lake-time-stamp",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}))

			Expect(stateStore.SetCall.CallCount).To(Equal(4))
			Expect(stateStore.SetCall.Receives[2].State).To(Equal(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
				KeyPair: storage.KeyPair{
					Name:       "keypair-bbl-lake-time-stamp",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				TFState: "some-tf-state",
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
					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
					Expect(err).To(MatchError("cannot apply"))

					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State).To(Equal(storage.State{
						TFState: "some-partial-tf-state",
					}))
				})

				Context("when the terraform manager error fails to return a bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.Error = errors.New("failed to retrieve bbl state")
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute(commands.AWSUpConfig{}, storage.State{})
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
							{},
							{errors.New("failed to set bbl state")},
						}
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute(commands.AWSUpConfig{}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\ncannot apply,\nfailed to set bbl state"))
					})
				})
			})

			Context("when the terraform manager fails with non terraformManagerError", func() {
				It("returns the error", func() {
					terraformManager.ApplyCall.Returns.Error = errors.New("cannot apply")

					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
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

					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
					Expect(err).To(MatchError("failed to set the state"))
				})
			})
		})

		Context("when the no-director flag is provided", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true
			})

			It("does not create a bosh or cloud config", func() {
				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "new-aws-access-key-id",
					SecretAccessKey: "new-aws-secret-access-key",
					Region:          "new-aws-region",
					NoDirector:      true,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(keyPairManager.SyncCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[1].State.NoDirector).To(BeTrue())
			})

			Context("when a bbl environment exists with no bosh director", func() {
				It("does not create a bosh director on subsequent runs", func() {
					err := command.Execute(commands.AWSUpConfig{
						AccessKeyID:     "new-aws-access-key-id",
						SecretAccessKey: "new-aws-secret-access-key",
						Region:          "new-aws-region",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
					Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
					Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
					Expect(keyPairManager.SyncCall.CallCount).To(Equal(1))
					Expect(stateStore.SetCall.CallCount).To(Equal(4))
				})
			})

			Context("when a bbl environment exists with a bosh director", func() {
				It("fast fails before creating any infrastructure", func() {
					err := command.Execute(commands.AWSUpConfig{
						AccessKeyID:     "new-aws-access-key-id",
						SecretAccessKey: "new-aws-secret-access-key",
						Region:          "new-aws-region",
						NoDirector:      true,
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
				KeyPair: storage.KeyPair{
					Name:       "keypair-bbl-lake-time-stamp",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				EnvID:   "bbl-lake-time-stamp",
				TFState: "some-tf-state",
			}

			err := command.Execute(commands.AWSUpConfig{}, incomingState)
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

				err = command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
					OpsFilePath:     opsFilePath,
				}, storage.State{
					EnvID: "bbl-lake-time-stamp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateDirectorCall.Receives.State.BOSH.UserOpsFile).To(Equal("some-ops-file-contents"))
			})
		})

		Context("when bosh az is provided via --aws-bosh-az flag", func() {
			It("passes the bosh az to terraform", func() {
				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
					BOSHAZ:          "some-bosh-az",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.Receives.BBLState.Stack.BOSHAZ).To(Equal("some-bosh-az"))
			})

			Context("when a stack exists and the aws-bosh-az is provided and different", func() {
				It("returns an error message", func() {
					err := command.Execute(commands.AWSUpConfig{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-aws-region",
						BOSHAZ:          "other-bosh-az",
					}, storage.State{
						Stack: storage.Stack{
							Name:   "some-stack",
							BOSHAZ: "some-bosh-az",
						},
					})
					Expect(err).To(MatchError("The --aws-bosh-az cannot be changed for existing environments."))
				})
			})
		})

		Describe("cloud config", func() {
			It("updates the bosh director with a cloud config provided an up-to-date state", func() {
				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(storage.State{
					EnvID: "bbl-lake-time-stamp",
					IAAS:  "aws",
					KeyPair: storage.KeyPair{
						Name:       "keypair-bbl-lake-time-stamp",
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
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
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-aws-region",
					},
					TFState: "some-tf-state",
				}))
			})
		})

		Describe("reentrant", func() {
			Context("when the key pair fails to sync", func() {
				It("saves the keypair name and returns an error", func() {
					keyPairManager.SyncCall.Returns.Error = keypair.NewManagerError(storage.State{
						KeyPair: storage.KeyPair{
							Name: "keypair-bbl-lake-time-stamp",
						},
					}, errors.New("error syncing key pair"))

					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
					Expect(err).To(MatchError("error syncing key pair"))
					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State.KeyPair.Name).To(Equal("keypair-bbl-lake-time-stamp"))
				})

				Context("when it can't save the state", func() {
					It("returns an error", func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set")}}
						keyPairManager.SyncCall.Returns.Error = keypair.NewManagerError(storage.State{
							KeyPair: storage.KeyPair{
								Name: "keypair-bbl-lake-time-stamp",
							},
						}, errors.New("error syncing key pair"))

						err := command.Execute(commands.AWSUpConfig{}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\nerror syncing key pair,\nfailed to set"))
						Expect(stateStore.SetCall.CallCount).To(Equal(2))
						Expect(stateStore.SetCall.Receives[1].State.KeyPair.Name).To(Equal("keypair-bbl-lake-time-stamp"))
					})
				})
			})
		})

		Describe("state manipulation", func() {
			Context("iaas", func() {
				It("writes iaas aws to state", func() {
					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives[3].State.IAAS).To(Equal("aws"))
				})
			})

			Context("aws credentials", func() {
				Context("when the credentials do not exist", func() {
					It("saves the credentials", func() {
						err := command.Execute(commands.AWSUpConfig{
							AccessKeyID:     "some-aws-access-key-id",
							SecretAccessKey: "some-aws-secret-access-key",
							Region:          "some-aws-region",
						}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.CallCount).To(Equal(5))
						Expect(stateStore.SetCall.Receives[1].State.AWS).To(Equal(storage.AWS{
							AccessKeyID:     "some-aws-access-key-id",
							SecretAccessKey: "some-aws-secret-access-key",
							Region:          "some-aws-region",
						}))
					})
				})

				Context("when the credentials do exist", func() {
					It("overrides the credentials when they're passed in", func() {
						err := command.Execute(commands.AWSUpConfig{
							AccessKeyID:     "new-aws-access-key-id",
							SecretAccessKey: "new-aws-secret-access-key",
							Region:          "new-aws-region",
						}, storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "old-aws-access-key-id",
								SecretAccessKey: "old-aws-secret-access-key",
								Region:          "old-aws-region",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives[1].State.AWS).To(Equal(storage.AWS{
							AccessKeyID:     "new-aws-access-key-id",
							SecretAccessKey: "new-aws-secret-access-key",
							Region:          "new-aws-region",
						}))
					})

					It("does not override the credentials when they're not passed in", func() {
						err := command.Execute(commands.AWSUpConfig{}, storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "aws-access-key-id",
								SecretAccessKey: "aws-secret-access-key",
								Region:          "aws-region",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives[1].State.AWS).To(Equal(storage.AWS{
							AccessKeyID:     "aws-access-key-id",
							SecretAccessKey: "aws-secret-access-key",
							Region:          "aws-region",
						}))
					})
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the env id manager fails", func() {
				envIDManager.SyncCall.Returns.Error = errors.New("env id manager failed")

				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
				}, storage.State{})
				Expect(err).To(MatchError("env id manager failed"))

			})

			It("returns an error when saving the state fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{
					{
						Error: errors.New("saving the state failed"),
					},
				}
				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
				}, storage.State{})
				Expect(err).To(MatchError("saving the state failed"))
			})

			It("returns an error when the cloud config cannot be uploaded", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update")
				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to update"))
			})

			It("returns an error when the broken environment validator fails", func() {
				brokenEnvironmentValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
				err := command.Execute(commands.AWSUpConfig{}, storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						Region: "some-aws-region",
					},
					BOSH: storage.BOSH{
						DirectorAddress: "some-director-address",
					},
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
				})

				Expect(brokenEnvironmentValidator.ValidateCall.Receives.State).To(Equal(storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						Region: "some-aws-region",
					},
					BOSH: storage.BOSH{
						DirectorAddress: "some-director-address",
					},
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
				}))

				Expect(err).To(MatchError("failed to validate"))

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(0))
			})

			It("returns an error when the terraform manager cannot get terraform outputs", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("cannot parse terraform output")

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("cannot parse terraform output"))
			})

			It("returns an error when the ops file cannot be read", func() {
				err := command.Execute(commands.AWSUpConfig{
					OpsFilePath: "some/fake/path",
				}, storage.State{})
				Expect(err).To(MatchError("open some/fake/path: no such file or directory"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshManager.CreateDirectorCall.Returns.Error = errors.New("cannot deploy bosh")

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("cannot deploy bosh"))
			})

			It("returns an error when state store fails to set the state before syncing the keypair", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before retrieving availability zones", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before updating the cloud config", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when only some of the AWS parameters are provided", func() {
				err := command.Execute(commands.AWSUpConfig{AccessKeyID: "some-key-id", Region: "some-region"}, storage.State{})
				Expect(err).To(MatchError("AWS secret access key must be provided"))
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
					err := command.Execute(commands.AWSUpConfig{}, incomingState)
					Expect(err).To(MatchError("failed to create"))
					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives[3].State.BOSH.State).To(Equal(expectedBOSHState))
				})

				It("returns a compound error when it fails to save the state", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("state failed to be set")}}
					err := command.Execute(commands.AWSUpConfig{}, incomingState)
					Expect(err).To(MatchError("the following errors occurred:\nfailed to create,\nstate failed to be set"))
					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives[3].State.BOSH.State).To(Equal(expectedBOSHState))
				})
			})
		})
	})
})
