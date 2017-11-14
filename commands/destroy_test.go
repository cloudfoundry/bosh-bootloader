package commands_test

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Destroy", func() {
	var (
		destroy                  commands.Destroy
		boshManager              *fakes.BOSHManager
		logger                   *fakes.Logger
		plan                     *fakes.Plan
		stateStore               *fakes.StateStore
		stateValidator           *fakes.StateValidator
		terraformManager         *fakes.TerraformManager
		networkDeletionValidator *fakes.NetworkDeletionValidator
		stdin                    *bytes.Buffer
	)

	BeforeEach(func() {
		stdin = bytes.NewBuffer([]byte{})
		logger = &fakes.Logger{}

		plan = &fakes.Plan{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"
		stateStore = &fakes.StateStore{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}
		networkDeletionValidator = &fakes.NetworkDeletionValidator{}

		// Returning a fully empty State is unrealistic.
		terraformManager.DestroyCall.Returns.BBLState = storage.State{ID: "some-state-id"}

		destroy = commands.NewDestroy(plan, logger, stdin, boshManager, stateStore,
			stateValidator, terraformManager, networkDeletionValidator)
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

		Context("if validating terraform version returns an error", func() {
			BeforeEach(func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("failed to validate version")
			})

			It("fast fails", func() {
				err := destroy.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to validate version"))
			})
		})

		Context("when there is no state and --skip-if-missing flag is provided", func() {
			It("returns no error", func() {
				err := destroy.CheckFastFails([]string{"--skip-if-missing"}, storage.State{})

				Expect(err).NotTo(HaveOccurred())
				Expect(logger.StepCall.Receives.Message).To(Equal("state file not found, and --skip-if-missing flag provided, exiting"))
			})
		})

		Context("when state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			})

			It("returns an error", func() {
				err := destroy.CheckFastFails([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})
		})

		Context("when iaas is gcp", func() {
			var bblState storage.State

			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
					"external_ip":        "some-external-ip",
					"network_name":       "some-network-name",
					"subnetwork_name":    "some-subnetwork-name",
					"bosh_open_tag_name": "some-bosh-tag",
					"internal_tag_name":  "some-internal-tag",
					"director_address":   "some-director-address",
				}}

				bblState = storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
				}
				terraformManager.DestroyCall.Returns.BBLState = bblState
			})

			Context("when there is no network name in the state", func() {
				It("does not attempt to validate whether it is safe to delete the network", func() {
					stdin.Write([]byte("yes\n"))
					terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{}

					err := destroy.CheckFastFails([]string{}, bblState)
					Expect(err).NotTo(HaveOccurred())

					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
				})
			})

			Context("when instances exist in the gcp network", func() {
				BeforeEach(func() {
					networkDeletionValidator.ValidateSafeToDeleteCall.Returns.Error = errors.New("validation failed")
				})

				It("returns an error", func() {
					err := destroy.CheckFastFails([]string{}, bblState)
					Expect(err).To(MatchError("validation failed"))
					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.Receives.NetworkName).To(Equal("some-network-name"))
				})
			})

			Context("when terraform output provider fails to get terraform outputs", func() {
				It("does not fast fail", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("terraform output provider failed")

					stdin.Write([]byte("yes\n"))
					err := destroy.CheckFastFails([]string{}, bblState)
					Expect(err).NotTo(HaveOccurred())
					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("when iaas is aws", func() {
			var state storage.State

			BeforeEach(func() {
				stdin.Write([]byte("yes\n"))
				state = storage.State{
					IAAS:  "aws",
					EnvID: "some-env-id",
				}
			})

			Context("if BOSH deployed VMs still exist in the VPC", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{
						Map: map[string]interface{}{"vpc_id": "some-vpc-id"},
					}
					networkDeletionValidator.ValidateSafeToDeleteCall.Returns.Error = errors.New("vpc some-vpc-id is not safe to delete")
				})

				It("fails fast", func() {
					err := destroy.CheckFastFails([]string{}, state)
					Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete"))

					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.Receives.NetworkName).To(Equal("some-vpc-id"))
					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.Receives.EnvID).To(Equal("some-env-id"))
				})
			})

			Context("when terraform manager fails to get outputs", func() {
				It("does not fast fail", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get outputs")

					err := destroy.CheckFastFails([]string{}, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("when iaas is azure", func() {
			// Replace this test with the pended test below when Azure supports ValidateSafeToDelete
			It("returns an error while instances exist in the azure network", func() {
				networkDeletionValidator.ValidateSafeToDeleteCall.Returns.Error = errors.New("validation failed")
				err := destroy.CheckFastFails([]string{}, storage.State{
					IAAS: "azure",
				})
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when instances exist in the azure network", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{
						Map: map[string]interface{}{"bosh_network_name": "some-network-id"},
					}
					networkDeletionValidator.ValidateSafeToDeleteCall.Returns.Error = errors.New("validation failed")
				})

				It("returns an error", func() {
					err := destroy.CheckFastFails([]string{}, storage.State{
						IAAS:  "azure",
						EnvID: "some-env-id",
					})
					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.Receives.NetworkName).To(Equal("some-network-id"))
					Expect(err).To(MatchError("validation failed"))
				})
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			plan.IsInitializedCall.Returns.IsInitialized = true
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
					Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(1))
				} else {
					Expect(logger.StepCall.Receives.Message).To(Equal("exiting"))
					Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(0))
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
				Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(1))
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

			Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.DeleteDirectorCall.Receives.State).To(Equal(state))

			Expect(stateStore.SetCall.CallCount).To(Equal(2))
			Expect(stateStore.SetCall.Receives[0].State.BOSH).To(Equal(storage.BOSH{}))
		})

		It("invokes bosh delete jumpbox as well", func() {
			stdin.Write([]byte("yes\n"))
			state := storage.State{
				BOSH: storage.BOSH{
					DirectorName: "some-director",
				},
				Jumpbox: storage.Jumpbox{
					Manifest: "some-manifest",
				},
			}
			stateWithoutDirector := storage.State{
				BOSH: storage.BOSH{},
				Jumpbox: storage.Jumpbox{
					Manifest: "some-manifest",
				},
			}

			err := destroy.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.IsInitializedCall.CallCount).To(Equal(1))
			Expect(plan.IsInitializedCall.Receives.State).To(Equal(state))
			Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.DeleteDirectorCall.Receives.State).To(Equal(state))
			Expect(boshManager.DeleteJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.DeleteJumpboxCall.Receives.State).To(Equal(stateWithoutDirector))

			Expect(stateStore.SetCall.CallCount).To(Equal(2))
			Expect(stateStore.SetCall.Receives[0].State.BOSH).To(Equal(storage.BOSH{}))
		})

		Context("when the plan is not initialized", func() {
			It("initializes the plan", func() {
				plan.IsInitializedCall.Returns.IsInitialized = false
				stdin.Write([]byte("yes\n"))
				state := storage.State{
					EnvID: "unintialized",
					BOSH: storage.BOSH{
						UserOpsFile: "some ops file contents",
					},
					NoDirector: true,
					LB:         storage.LB{Type: "lb-type", Domain: "lb-domain"},
				}
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(plan.IsInitializedCall.CallCount).To(Equal(1))
				Expect(plan.IsInitializedCall.Receives.State).To(Equal(state))
				Expect(plan.InitializePlanCall.CallCount).To(Equal(1))
				Expect(plan.InitializePlanCall.Receives.State).To(Equal(state))
				Expect(plan.InitializePlanCall.Receives.Plan).To(Equal(commands.PlanConfig{
					Name:       "unintialized",
					OpsFile:    "some ops file contents",
					NoDirector: true,
					LB:         storage.LB{Type: "lb-type", Domain: "lb-domain"},
				}))
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
					boshManager.DeleteDirectorCall.Returns.Error = errors.New("bosh delete-env failed")

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
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						State: map[string]interface{}{
							"key": "value",
						},
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
					},
					EnvID: "bbl-lake-time:stamp",
				}
			})

			It("calls terraform destroy and deletes the state file", func() {
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				expectedState := state
				expectedState.BOSH = storage.BOSH{}
				Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(expectedState))
				Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(expectedState))
				Expect(stateStore.SetCall.Receives[1].State).To(Equal(storage.State{}))
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

					terraformManager.DestroyCall.Returns.BBLState = updatedBBLState
					terraformManager.DestroyCall.Returns.Error = errors.New("failed to destroy")

					stdin.Write([]byte("yes\n"))
				})

				It("saves the partially destroyed tf state", func() {
					err := destroy.Execute([]string{}, state)
					Expect(err).To(MatchError("failed to destroy"))

					Expect(terraformManager.InitCall.CallCount).To(Equal(1))
					Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(expectedBBLState))
					Expect(terraformManager.DestroyCall.CallCount).To(Equal(1))
					Expect(terraformManager.DestroyCall.Receives.BBLState).To(Equal(expectedBBLState))

					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State).To(Equal(updatedBBLState))
				})

				Context("when the state fails to be set", func() {
					It("returns an error containing both messages", func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}
						err := destroy.Execute([]string{}, storage.State{
							IAAS: "gcp",
						})

						Expect(err).To(MatchError("the following errors occurred:\nfailed to destroy,\nfailed to set state"))
					})
				})
			})

			It("logs the bosh deletion", func() {
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.StepCall.Messages).To(ContainElement("destroying bosh director"))
			})

			Context("reentrance", func() {
				Context("when NoDirector is true", func() {
					It("does not attempt to delete the bosh director", func() {
						state.NoDirector = true
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintlnCall.Receives.Message).To(Equal("no BOSH director, skipping..."))
						Expect(logger.StepCall.Messages).NotTo(ContainElement("destroying bosh director"))
						Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(0))
					})
				})

				Context("when state.BOSH is empty", func() {
					It("does not attempt to delete the bosh director", func() {
						state.BOSH = storage.BOSH{}
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintlnCall.Receives.Message).NotTo(Equal("no BOSH director, skipping..."))
						Expect(logger.StepCall.Messages).NotTo(ContainElement("destroying bosh director"))
						Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(0))
					})
				})

				Context("when state.Jumpbox is empty", func() {
					It("does not attempt to delete the jumpbox", func() {
						state.Jumpbox = storage.Jumpbox{}
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.StepCall.Messages).To(ContainElement("destroying bosh director"))
						Expect(logger.StepCall.Messages).NotTo(ContainElement("destroying jumpbox"))
						Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(1))
						Expect(boshManager.DeleteJumpboxCall.CallCount).To(Equal(0))
					})
				})
			})
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				stdin.Write([]byte("yes\n"))
			})

			Context("when bosh fails to delete the director", func() {
				var state storage.State

				BeforeEach(func() {
					state = storage.State{
						BOSH: storage.BOSH{
							State: map[string]interface{}{"hello": "world"},
						},
						IAAS: "aws",
					}
				})
				Context("when bosh delete returns a bosh manager delete error", func() {
					var errState storage.State

					BeforeEach(func() {
						errState = storage.State{
							BOSH: storage.BOSH{
								State: map[string]interface{}{"error": "state"},
							},
							IAAS: "aws",
						}
						boshManager.DeleteDirectorCall.Returns.Error = bosh.NewManagerDeleteError(errState, errors.New("deletion failed"))
					})

					It("saves the bosh state and returns an error", func() {
						err := destroy.Execute([]string{}, state)
						Expect(err).To(MatchError("deletion failed"))
						Expect(stateStore.SetCall.CallCount).To(Equal(1))
						Expect(stateStore.SetCall.Receives[0].State).To(Equal(errState))
					})

					Context("when it can't set the state", func() {
						BeforeEach(func() {
							stateStore.SetCall.Returns = []fakes.SetCallReturn{{
								errors.New("saving state failed"),
							}}
						})
						It("returns an error", func() {
							err := destroy.Execute([]string{}, state)
							Expect(err).To(MatchError("the following errors occurred:\ndeletion failed,\nsaving state failed"))
						})
					})
				})

				It("returns an error", func() {
					boshManager.DeleteDirectorCall.Returns.Error = errors.New("deletion failed")
					err := destroy.Execute([]string{}, state)
					Expect(err).To(MatchError("deletion failed"))
				})
			})
		})
	})
})
