package commands_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Destroy", func() {
	var (
		destroy                 commands.Destroy
		boshManager             *fakes.BOSHManager
		stackManager            *fakes.StackManager
		infrastructureManager   *fakes.InfrastructureManager
		vpcStatusChecker        *fakes.VPCStatusChecker
		logger                  *fakes.Logger
		certificateDeleter      *fakes.CertificateDeleter
		stateStore              *fakes.StateStore
		stateValidator          *fakes.StateValidator
		terraformManager        *fakes.TerraformManager
		terraformManagerError   *fakes.TerraformManagerError
		networkInstancesChecker *fakes.NetworkInstancesChecker
		stdin                   *bytes.Buffer
	)

	BeforeEach(func() {
		stdin = bytes.NewBuffer([]byte{})
		logger = &fakes.Logger{}

		vpcStatusChecker = &fakes.VPCStatusChecker{}
		stackManager = &fakes.StackManager{}
		infrastructureManager = &fakes.InfrastructureManager{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"
		certificateDeleter = &fakes.CertificateDeleter{}
		stateStore = &fakes.StateStore{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}
		terraformManagerError = &fakes.TerraformManagerError{}
		networkInstancesChecker = &fakes.NetworkInstancesChecker{}

		destroy = commands.NewDestroy(logger, stdin, boshManager,
			vpcStatusChecker, stackManager, infrastructureManager,
			certificateDeleter, stateStore,
			stateValidator, terraformManager, networkInstancesChecker)
	})

	Describe("CheckFastFails", func() {
		Context("when the BOSH version is less than 2.0.24 and there is a director", func() {
			It("returns a helpful error message", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := destroy.CheckFastFails([]string{"--skip-if-missing"}, storage.State{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})
		})

		Context("when the BOSH version is less than 2.0.24 and there is no director", func() {
			It("does not fast fail", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := destroy.CheckFastFails([]string{"--skip-if-missing"}, storage.State{
					IAAS:       "aws",
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("fast fails on gcp if the terraform installed is less than v0.8.5", func() {
			terraformManager.ValidateVersionCall.Returns.Error = errors.New("failed to validate version")

			err := destroy.CheckFastFails([]string{}, storage.State{IAAS: "gcp"})
			Expect(err).To(MatchError("failed to validate version"))
		})

		It("does not fast fail on aws if the terraform installed is less than v0.8.5", func() {
			err := destroy.CheckFastFails([]string{}, storage.State{IAAS: "aws"})
			Expect(err).ToNot(HaveOccurred())
			Expect(terraformManager.ValidateVersionCall.CallCount).To(Equal(0))
		})

		It("returns when there is no state and --skip-if-missing flag is provided", func() {
			err := destroy.CheckFastFails([]string{"--skip-if-missing"}, storage.State{})

			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Receives.Message).To(Equal("state file not found, and --skip-if-missing flag provided, exiting"))
		})

		It("returns an error when state validator fails", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			err := destroy.CheckFastFails([]string{}, storage.State{})

			Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
			Expect(err).To(MatchError("state validator failed"))
		})

		Context("when iaas is gcp", func() {
			var (
				serviceAccountKeyPath string
				serviceAccountKey     string
				bblState              storage.State
			)

			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"external_ip":        "some-external-ip",
					"network_name":       "some-network-name",
					"subnetwork_name":    "some-subnetwork-name",
					"bosh_open_tag_name": "some-bosh-tag",
					"internal_tag_name":  "some-internal-tag",
					"director_address":   "some-director-address",
				}

				tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
				Expect(err).NotTo(HaveOccurred())
				serviceAccountKeyPath = tempFile.Name()
				serviceAccountKey = `{"real": "json"}`
				err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				bblState = storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					TFState: "some-tf-state",
					KeyPair: storage.KeyPair{
						PublicKey: "some-public-key",
					},
				}
				terraformManager.DestroyCall.Returns.BBLState = bblState
			})

			Context("when there is no network name in the state", func() {
				It("does not attempt to validate if it is safe to delete the network", func() {
					stdin.Write([]byte("yes\n"))
					terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{}

					err := destroy.CheckFastFails([]string{}, storage.State{
						IAAS: "gcp",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(networkInstancesChecker.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
				})
			})

			It("returns an error when instances exist in the gcp network", func() {
				networkInstancesChecker.ValidateSafeToDeleteCall.Returns.Error = errors.New("validation failed")

				projectID := "some-project-id"
				zone := "some-zone"
				tfState := "some-tf-state"
				err := destroy.CheckFastFails([]string{}, storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         projectID,
						Zone:              zone,
						Region:            "some-region",
					},
					TFState: tfState,
				})

				Expect(networkInstancesChecker.ValidateSafeToDeleteCall.Receives.NetworkName).To(Equal("some-network-name"))
				Expect(err).To(MatchError("validation failed"))
			})

			Context("when terraform output provider fails to get terraform outputs", func() {
				It("does not fast fail", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("terraform output provider failed")

					stdin.Write([]byte("yes\n"))
					err := destroy.CheckFastFails([]string{}, storage.State{
						IAAS: "gcp",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(networkInstancesChecker.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("when iaas is aws", func() {
			var (
				state storage.State
			)

			BeforeEach(func() {
				stdin.Write([]byte("yes\n"))
				state = storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-aws-region",
					},
					KeyPair: storage.KeyPair{
						Name:       "some-ec2-key-pair-name",
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
					},
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						State: map[string]interface{}{
							"key": "value",
						},
						Credentials: map[string]string{
							"some-username": "some-password",
						},
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
					},
					EnvID: "bbl-lake-time:stamp",
				}
			})

			Context("when cloudformation was used", func() {
				BeforeEach(func() {
					state.Stack = storage.Stack{
						Name:            "some-stack-name",
						LBType:          "some-lb-type",
						CertificateName: "some-certificate-name",
					}
				})

				It("fails fast if BOSH deployed VMs still exist in the VPC with cloudformation", func() {
					stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Name:   "some-stack-name",
						Status: "some-stack-status",
						Outputs: map[string]string{
							"VPCID": "some-vpc-id",
						},
					}
					vpcStatusChecker.ValidateSafeToDeleteCall.Returns.Error = errors.New("vpc some-vpc-id is not safe to delete")

					err := destroy.CheckFastFails([]string{}, state)
					Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete"))

					Expect(vpcStatusChecker.ValidateSafeToDeleteCall.Receives.VPCID).To(Equal("some-vpc-id"))
					Expect(vpcStatusChecker.ValidateSafeToDeleteCall.Receives.EnvID).To(Equal(""))
				})

				Context("when the stack manager cannot describe the stack", func() {
					It("returns an error", func() {
						stackManager.DescribeCall.Returns.Error = errors.New("cannot describe stack")

						err := destroy.CheckFastFails([]string{}, storage.State{
							IAAS: "aws",
						})
						Expect(err).To(MatchError("cannot describe stack"))
					})
				})
			})

			Context("when terraform was used", func() {
				BeforeEach(func() {
					state.TFState = "some-tf-state"
					state.EnvID = "some-env-id"
				})

				Context("when terraform manager fails to get outputs", func() {
					It("does not fast fail", func() {
						terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get outputs")

						err := destroy.CheckFastFails([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
						Expect(vpcStatusChecker.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
					})
				})

				It("fails fast if BOSH deployed VMs still exist in the VPC", func() {
					terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
						"vpc_id": "some-vpc-id",
					}

					vpcStatusChecker.ValidateSafeToDeleteCall.Returns.Error = errors.New("vpc some-vpc-id is not safe to delete")
					Expect(state.Stack.Name).To(BeEmpty())

					err := destroy.CheckFastFails([]string{}, state)
					Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete"))

					Expect(vpcStatusChecker.ValidateSafeToDeleteCall.Receives.VPCID).To(Equal("some-vpc-id"))
					Expect(vpcStatusChecker.ValidateSafeToDeleteCall.Receives.EnvID).To(Equal("some-env-id"))
				})
			})
		})
	})

	Describe("Execute", func() {
		It("returns when there is no state and --skip-if-missing flag is provided", func() {
			err := destroy.Execute([]string{"--skip-if-missing"}, storage.State{})

			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Receives.Message).To(Equal("state file not found, and --skip-if-missing flag provided, exiting"))
		})

		DescribeTable("prompting the user for confirmation",
			func(response string, proceed bool) {
				fmt.Fprintf(stdin, "%s\n", response)

				err := destroy.Execute([]string{}, storage.State{
					BOSH: storage.BOSH{
						DirectorName: "some-director",
					},
					EnvID: "some-lake",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal(`Are you sure you want to delete infrastructure for "some-lake"? This operation cannot be undone!`))

				if proceed {
					Expect(boshManager.DeleteCall.CallCount).To(Equal(1))
				} else {
					Expect(logger.StepCall.Receives.Message).To(Equal("exiting"))
					Expect(boshManager.DeleteCall.CallCount).To(Equal(0))
				}
			},
			Entry("responding with 'yes'", "yes", true),
			Entry("responding with 'y'", "y", true),
			Entry("responding with 'Yes'", "Yes", true),
			Entry("responding with 'Y'", "Y", true),
			Entry("responding with 'no'", "no", false),
			Entry("responding with 'n'", "n", false),
			Entry("responding with 'No'", "No", false),
			Entry("responding with 'N'", "N", false),
		)

		Context("when the --no-confirm flag is supplied", func() {
			DescribeTable("destroys without prompting the user for confirmation", func(flag string) {
				err := destroy.Execute([]string{flag}, storage.State{
					BOSH: storage.BOSH{
						DirectorName: "some-director",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(boshManager.DeleteCall.CallCount).To(Equal(1))
			},
				Entry("--no-confirm", "--no-confirm"),
				Entry("-n", "-n"),
			)
		})

		It("invokes bosh delete", func() {
			stdin.Write([]byte("yes\n"))
			state := storage.State{
				BOSH: storage.BOSH{
					DirectorName: "some-director",
				},
			}

			err := destroy.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshManager.DeleteCall.CallCount).To(Equal(1))
			Expect(boshManager.DeleteCall.Receives.State).To(Equal(state))
		})

		It("clears the state", func() {
			stdin.Write([]byte("yes\n"))
			err := destroy.Execute([]string{}, storage.State{
				Stack: storage.Stack{
					Name:            "some-stack-name",
					LBType:          "some-lb-type",
					CertificateName: "some-certificate-name",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(stateStore.SetCall.CallCount).To(Equal(3))
			Expect(stateStore.SetCall.Receives[2].State).To(Equal(storage.State{}))
		})

		Context("when jumpbox is enabled", func() {
			It("invokes bosh delete jumpbox as well", func() {
				stdin.Write([]byte("yes\n"))
				state := storage.State{
					BOSH: storage.BOSH{
						DirectorName: "some-director",
					},
					Jumpbox: storage.Jumpbox{
						Enabled:  true,
						Manifest: "some-manifest",
					},
				}
				stateWithoutDirector := storage.State{
					BOSH: storage.BOSH{},
					Jumpbox: storage.Jumpbox{
						Enabled:  true,
						Manifest: "some-manifest",
					},
				}

				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.DeleteCall.CallCount).To(Equal(1))
				Expect(boshManager.DeleteCall.Receives.State).To(Equal(state))
				Expect(boshManager.DeleteJumpboxCall.CallCount).To(Equal(1))
				Expect(boshManager.DeleteJumpboxCall.Receives.State).To(Equal(stateWithoutDirector))
			})

			It("clears the state", func() {
				stdin.Write([]byte("yes\n"))
				err := destroy.Execute([]string{}, storage.State{
					BOSH:    storage.BOSH{},
					Jumpbox: storage.Jumpbox{},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(stateStore.SetCall.CallCount).To(Equal(3))
				Expect(stateStore.SetCall.Receives[2].State).To(Equal(storage.State{}))
			})
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				stdin.Write([]byte("yes\n"))
			})

			Context("when an invalid command line flag is supplied", func() {
				It("returns an error", func() {
					err := destroy.Execute([]string{"--invalid-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
				})
			})

			Context("when the terraform manager fails to get outputs", func() {
				It("returns an error", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("nope")

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("nope"))
				})
			})

			Context("when bosh delete fails", func() {
				It("returns an error", func() {
					boshManager.DeleteCall.Returns.Error = errors.New("bosh delete-env failed")

					err := destroy.Execute([]string{}, storage.State{
						BOSH: storage.BOSH{
							DirectorName: "some-director",
						},
					})
					Expect(err).To(MatchError("bosh delete-env failed"))
				})
			})

			Context("when state store fails to set the state before destroying infrastructure", func() {
				It("returns an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set state")}}

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to set state"))
				})
			})

			Context("when state store fails to set the state before destroying certificate", func() {
				It("returns an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to set state"))
				})
			})

			Context("when the state fails to be set", func() {
				It("return an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("failed to set state")}}

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to set state"))
				})
			})
		})

		Context("when iaas is azure", func() {
			var (
				state        storage.State
				updatedState storage.State
			)

			BeforeEach(func() {
				stdin.Write([]byte("yes\n"))
				state = storage.State{
					IAAS: "azure",
				}

				updatedState = storage.State{
					IAAS:    "azure",
					TFState: "some-tf-state",
				}
				terraformManager.DestroyCall.Returns.BBLState = updatedState
			})

			It("deletes infrastructure with terraform", func() {
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(state))
				Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(state))
			})

			Context("when terraform destroy fails", func() {
				BeforeEach(func() {
					terraformManagerError.ErrorCall.Returns = "failed to destroy"
					terraformManagerError.BBLStateCall.Returns.BBLState = updatedState

					terraformManager.DestroyCall.Returns.BBLState = storage.State{}
					terraformManager.DestroyCall.Returns.Error = terraformManagerError

					stdin.Write([]byte("yes\n"))
				})

				It("saves the partially destroyed tf state", func() {
					err := destroy.Execute([]string{}, state)
					Expect(err).To(Equal(terraformManagerError))

					Expect(terraformManager.DestroyCall.CallCount).To(Equal(1))
					Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(state))

					Expect(terraformManagerError.BBLStateCall.CallCount).To(Equal(1))

					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State).To(Equal(updatedState))
				})

				Context("when we cannot retrieve the updated bbl state", func() {
					BeforeEach(func() {
						terraformManagerError.BBLStateCall.Returns.Error = errors.New("some-bbl-state-error")
					})

					It("returns an error containing both messages", func() {
						err := destroy.Execute([]string{}, state)

						Expect(err).To(MatchError("the following errors occurred:\nfailed to destroy,\nsome-bbl-state-error"))
						Expect(stateStore.SetCall.CallCount).To(Equal(1))
					})
				})

				Context("and the state fails to be set", func() {
					It("returns an error containing both messages", func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}
						err := destroy.Execute([]string{}, storage.State{
							IAAS: "azure",
						})

						Expect(err).To(MatchError("the following errors occurred:\nfailed to destroy,\nfailed to set state"))
					})
				})
			})
		})

		Context("when iaas is aws", func() {
			Describe("destroying the aws infrastructure", func() {
				var (
					state storage.State
				)

				BeforeEach(func() {
					stdin.Write([]byte("yes\n"))
					state = storage.State{
						IAAS: "aws",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-aws-region",
						},
						KeyPair: storage.KeyPair{
							Name:       "some-ec2-key-pair-name",
							PrivateKey: "some-private-key",
							PublicKey:  "some-public-key",
						},
						BOSH: storage.BOSH{
							DirectorUsername: "some-director-username",
							DirectorPassword: "some-director-password",
							State: map[string]interface{}{
								"key": "value",
							},
							Credentials: map[string]string{
								"some-username": "some-password",
							},
							DirectorSSLCertificate: "some-certificate",
							DirectorSSLPrivateKey:  "some-private-key",
						},
						Stack: storage.Stack{
							Name:            "some-stack-name",
							LBType:          "some-lb-type",
							CertificateName: "some-certificate-name",
						},
						EnvID: "bbl-lake-time:stamp",
					}
				})

				Context("when infrastructure was created with cloudformation", func() {
					It("deletes the stack", func() {
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.StepCall.Messages).To(ContainElement("destroying AWS stack"))
						Expect(infrastructureManager.DeleteCall.Receives.StackName).To(Equal("some-stack-name"))
					})
				})

				Context("when infrastructure was created with terraform", func() {
					BeforeEach(func() {
						state.Stack = storage.Stack{}
						state.TFState = "some-tf-state"
						state.EnvID = "some-env-id"
					})

					It("deletes infrastructure with terraform", func() {
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(state))
						expectedState := state
						expectedState.BOSH = storage.BOSH{}
						Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(expectedState))
					})

					Context("when terraform destroy fails", func() {
						var (
							expectedBBLState storage.State
							updatedBBLState  storage.State
						)

						BeforeEach(func() {
							expectedBBLState = state
							expectedBBLState.BOSH = storage.BOSH{}

							updatedBBLState = state
							updatedBBLState.TFState = "some-updated-tf-state"

							terraformManagerError.ErrorCall.Returns = "failed to destroy"
							terraformManagerError.BBLStateCall.Returns.BBLState = updatedBBLState

							terraformManager.DestroyCall.Returns.BBLState = storage.State{}
							terraformManager.DestroyCall.Returns.Error = terraformManagerError

							stdin.Write([]byte("yes\n"))
						})

						It("saves the partially destroyed tf state", func() {
							err := destroy.Execute([]string{}, state)
							Expect(err).To(Equal(terraformManagerError))

							Expect(terraformManager.DestroyCall.CallCount).To(Equal(1))
							Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(expectedBBLState))

							Expect(terraformManagerError.BBLStateCall.CallCount).To(Equal(1))

							Expect(stateStore.SetCall.CallCount).To(Equal(2))
							Expect(stateStore.SetCall.Receives[1].State).To(Equal(updatedBBLState))
						})

						Context("when we cannot retrieve the updated bbl state", func() {
							BeforeEach(func() {
								terraformManagerError.BBLStateCall.Returns.Error = errors.New("some-bbl-state-error")
							})

							It("returns an error containing both messages", func() {
								err := destroy.Execute([]string{}, state)

								Expect(err).To(MatchError("the following errors occurred:\nfailed to destroy,\nsome-bbl-state-error"))
								Expect(stateStore.SetCall.CallCount).To(Equal(1))
							})
						})

						Context("and the state fails to be set", func() {
							It("returns an error containing both messages", func() {
								stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}
								err := destroy.Execute([]string{}, storage.State{
									IAAS: "gcp",
								})

								Expect(err).To(MatchError("the following errors occurred:\nfailed to destroy,\nfailed to set state"))
							})
						})
					})
				})

				It("deletes the certificate", func() {
					err := destroy.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(certificateDeleter.DeleteCall.Receives.CertificateName).To(Equal("some-certificate-name"))
					Expect(logger.StepCall.Messages).To(ContainElement("deleting certificate"))
				})

				Context("when there is no certficate to delete", func() {
					It("doesn't call delete certificate", func() {
						state.Stack.CertificateName = ""
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(certificateDeleter.DeleteCall.CallCount).To(Equal(0))
					})
				})

				It("logs the bosh deletion", func() {
					err := destroy.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.StepCall.Messages).To(ContainElement("destroying bosh director"))
				})

				Context("reentrance", func() {
					Context("when the stack fails to delete", func() {
						It("removes the bosh properties from state and returns an error", func() {
							infrastructureManager.DeleteCall.Returns.Error = errors.New("failed to delete stack")

							err := destroy.Execute([]string{}, state)
							Expect(err).To(MatchError("failed to delete stack"))

							Expect(stateStore.SetCall.CallCount).To(Equal(1))
							Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
								IAAS: "aws",
								AWS: storage.AWS{
									AccessKeyID:     "some-access-key-id",
									SecretAccessKey: "some-secret-access-key",
									Region:          "some-aws-region",
								},
								KeyPair: storage.KeyPair{
									Name:       "some-ec2-key-pair-name",
									PrivateKey: "some-private-key",
									PublicKey:  "some-public-key",
								},
								BOSH: storage.BOSH{},
								Stack: storage.Stack{
									Name:            "some-stack-name",
									LBType:          "some-lb-type",
									CertificateName: "some-certificate-name",
								},
								EnvID: "bbl-lake-time:stamp",
							}))
						})
					})

					Context("when there is no bosh to delete", func() {
						It("does not attempt to delete the bosh", func() {
							state.NoDirector = true
							err := destroy.Execute([]string{}, state)
							Expect(err).NotTo(HaveOccurred())

							Expect(logger.PrintlnCall.Receives.Message).To(Equal("no BOSH director, skipping..."))
							Expect(logger.StepCall.Messages).NotTo(ContainElement("destroying bosh director"))
							Expect(boshManager.DeleteCall.CallCount).To(Equal(0))
						})
					})

					Context("when the certificate fails to delete", func() {
						It("removes the stack from the state and returns an error", func() {
							certificateDeleter.DeleteCall.Returns.Error = errors.New("failed to delete certificate")

							err := destroy.Execute([]string{}, state)
							Expect(err).To(MatchError("failed to delete certificate"))

							Expect(stateStore.SetCall.CallCount).To(Equal(2))
							Expect(stateStore.SetCall.Receives[1].State).To(Equal(storage.State{
								IAAS: "aws",
								AWS: storage.AWS{
									AccessKeyID:     "some-access-key-id",
									SecretAccessKey: "some-secret-access-key",
									Region:          "some-aws-region",
								},
								KeyPair: storage.KeyPair{
									Name:       "some-ec2-key-pair-name",
									PrivateKey: "some-private-key",
									PublicKey:  "some-public-key",
								},
								BOSH: storage.BOSH{},
								Stack: storage.Stack{
									Name:            "",
									LBType:          "",
									CertificateName: "some-certificate-name",
								},
								EnvID: "bbl-lake-time:stamp",
							}))
						})
					})

					Context("when there is no stack to delete", func() {
						BeforeEach(func() {
							stackManager.DescribeCall.Returns.Error = cloudformation.StackNotFound
						})

						It("does not validate the vpc", func() {
							state.Stack = storage.Stack{}
							err := destroy.Execute([]string{}, state)
							Expect(err).NotTo(HaveOccurred())

							Expect(vpcStatusChecker.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
						})

						It("does not attempt to delete the stack", func() {
							state.Stack = storage.Stack{}
							err := destroy.Execute([]string{}, state)
							Expect(err).NotTo(HaveOccurred())

							Expect(logger.PrintlnCall.Receives.Message).To(Equal("No infrastructure found, skipping..."))
							Expect(infrastructureManager.DeleteCall.CallCount).To(Equal(0))
						})
					})
				})
			})

			Context("failure cases", func() {
				BeforeEach(func() {
					stdin.Write([]byte("yes\n"))
				})

				Context("when the stack manager cannot describe the stack", func() {
					It("returns an error", func() {
						stackManager.DescribeCall.Returns.Error = errors.New("cannot describe stack")

						err := destroy.Execute([]string{}, storage.State{
							IAAS: "aws",
						})
						Expect(err).To(MatchError("cannot describe stack"))
					})
				})

				Context("when failing to delete the stack", func() {
					It("returns an error", func() {
						infrastructureManager.DeleteCall.Returns.Error = errors.New("failed to delete stack")

						err := destroy.Execute([]string{}, storage.State{
							IAAS: "aws",
							Stack: storage.Stack{
								Name: "some-stack-name",
							},
						})
						Expect(err).To(MatchError("failed to delete stack"))
					})
				})

				Context("when the certificate cannot be deleted", func() {
					It("returns an error", func() {
						certificateDeleter.DeleteCall.Returns.Error = errors.New("failed to delete certificate")

						err := destroy.Execute([]string{}, storage.State{
							IAAS: "aws",
							Stack: storage.Stack{
								CertificateName: "some-certificate",
							}})
						Expect(err).To(MatchError("failed to delete certificate"))
					})
				})

				Context("when bosh fails to delete the director", func() {
					var (
						state storage.State
					)
					BeforeEach(func() {
						state = storage.State{
							BOSH: storage.BOSH{
								State: map[string]interface{}{"hello": "world"},
							},
							IAAS: "aws",
							Stack: storage.Stack{
								CertificateName: "some-certificate-name",
							},
						}
					})
					Context("when bosh delete returns a bosh manager delete error", func() {
						var (
							errState storage.State
						)
						BeforeEach(func() {
							errState = storage.State{
								BOSH: storage.BOSH{
									State: map[string]interface{}{"error": "state"},
								},
								IAAS: "aws",
								Stack: storage.Stack{
									CertificateName: "some-certificate-name",
								},
							}
							boshManager.DeleteCall.Returns.Error = bosh.NewManagerDeleteError(errState, errors.New("deletion failed"))
						})

						It("saves the bosh state and returns an error", func() {
							err := destroy.Execute([]string{}, state)
							Expect(err).To(MatchError("deletion failed"))
							Expect(stateStore.SetCall.CallCount).To(Equal(1))
							Expect(stateStore.SetCall.Receives[0].State).To(Equal(errState))
						})

						It("returns an error when it can't set the state", func() {
							stateStore.SetCall.Returns = []fakes.SetCallReturn{{
								errors.New("saving state failed"),
							}}
							err := destroy.Execute([]string{}, state)
							Expect(err).To(MatchError("the following errors occurred:\ndeletion failed,\nsaving state failed"))
						})
					})

					It("returns an error", func() {
						boshManager.DeleteCall.Returns.Error = errors.New("deletion failed")
						err := destroy.Execute([]string{}, state)
						Expect(err).To(MatchError("deletion failed"))
					})
				})
			})
		})

		Context("when iaas is gcp", func() {
			var serviceAccountKeyPath string
			var serviceAccountKey string
			var bblState storage.State

			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"external_ip":        "some-external-ip",
					"network_name":       "some-network-name",
					"subnetwork_name":    "some-subnetwork-name",
					"bosh_open_tag_name": "some-bosh-tag",
					"internal_tag_name":  "some-internal-tag",
					"director_address":   "some-director-address",
				}

				tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
				Expect(err).NotTo(HaveOccurred())
				serviceAccountKeyPath = tempFile.Name()
				serviceAccountKey = `{"real": "json"}`
				err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				bblState = storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					TFState: "some-tf-state",
					KeyPair: storage.KeyPair{
						PublicKey: "some-public-key",
					},
				}
				terraformManager.DestroyCall.Returns.BBLState = bblState
			})

			It("calls terraform destroy", func() {
				stdin.Write([]byte("yes\n"))
				err := destroy.Execute([]string{}, bblState)
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(bblState))
				Expect(terraformManager.DestroyCall.CallCount).To(Equal(1))
				Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(bblState))
			})

			Context("when terraform destroy fails", func() {
				var (
					updatedBBLState storage.State
				)

				BeforeEach(func() {
					updatedBBLState = bblState
					updatedBBLState.TFState = "some-updated-tf-state"

					terraformManagerError.ErrorCall.Returns = "failed to destroy"
					terraformManagerError.BBLStateCall.Returns.BBLState = updatedBBLState

					terraformManager.DestroyCall.Returns.BBLState = storage.State{}
					terraformManager.DestroyCall.Returns.Error = terraformManagerError

					stdin.Write([]byte("yes\n"))
				})

				It("saves the partially destroyed tf state", func() {
					err := destroy.Execute([]string{}, bblState)
					Expect(err).To(Equal(terraformManagerError))

					Expect(terraformManager.DestroyCall.CallCount).To(Equal(1))
					Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(bblState))

					Expect(terraformManagerError.BBLStateCall.CallCount).To(Equal(1))

					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State).To(Equal(updatedBBLState))
				})

				Context("when we cannot retrieve the updated bbl state", func() {
					BeforeEach(func() {
						terraformManagerError.BBLStateCall.Returns.Error = errors.New("some-bbl-state-error")
					})

					It("returns an error containing both messages", func() {
						err := destroy.Execute([]string{}, bblState)

						Expect(err).To(MatchError("the following errors occurred:\nfailed to destroy,\nsome-bbl-state-error"))
						Expect(stateStore.SetCall.CallCount).To(Equal(1))
					})
				})

				Context("and the state fails to be set", func() {
					It("returns an error containing both messages", func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}
						err := destroy.Execute([]string{}, storage.State{
							IAAS: "gcp",
						})

						Expect(err).To(MatchError("the following errors occurred:\nfailed to destroy,\nfailed to set state"))
					})
				})
			})
		})
	})
})
