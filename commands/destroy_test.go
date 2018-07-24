package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Destroy", func() {
	var (
		destroy commands.Destroy

		boshManager              *fakes.BOSHManager
		logger                   *fakes.Logger
		plan                     *fakes.Plan
		stateStore               *fakes.StateStore
		stateValidator           *fakes.StateValidator
		terraformManager         *fakes.TerraformManager
		networkDeletionValidator *fakes.NetworkDeletionValidator
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		logger.PromptCall.Returns.Proceed = true

		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.48"
		boshManager.PathCall.Returns.Path = "/bin/bosh1"

		plan = &fakes.Plan{}
		stateStore = &fakes.StateStore{}
		stateValidator = &fakes.StateValidator{}
		networkDeletionValidator = &fakes.NetworkDeletionValidator{}

		terraformManager = &fakes.TerraformManager{}
		terraformManager.DestroyCall.Returns.BBLState = storage.State{ID: "some-state-id"}
		terraformManager.IsPavedCall.Returns.IsPaved = true

		destroy = commands.NewDestroy(plan, logger, boshManager, stateStore,
			stateValidator, terraformManager, networkDeletionValidator)
	})

	Describe("CheckFastFails", func() {
		Context("when the BOSH version is less than 2.0.48 and there is a director", func() {
			It("returns a helpful error message", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := destroy.CheckFastFails([]string{}, storage.State{IAAS: "aws"})
				Expect(err).To(MatchError("/bin/bosh1: bosh-cli version must be at least v2.0.48"))
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

		Context("when state validator doesn't find a bbl state", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = commands.NewNoBBLStateError("lol")
			})

			It("logs and indicates that we should exit sucessfully", func() {
				err := destroy.CheckFastFails([]string{}, storage.State{})

				Expect(logger.PrintlnCall.Receives.Message).To(ContainSubstring("bbl-state.json not found"))
				Expect(err).To(MatchError(commands.ExitSuccessfully{}))
			})
		})

		Context("when the environment is not paved", func() {
			It("deletes the directory without attempting to destroy bosh or terraform", func() {
				terraformManager.IsPavedCall.Returns.IsPaved = false

				err := destroy.CheckFastFails([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(0))
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

					err := destroy.CheckFastFails([]string{}, bblState)
					Expect(err).NotTo(HaveOccurred())
					Expect(networkDeletionValidator.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("when iaas is aws", func() {
			var state storage.State

			BeforeEach(func() {
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

		It("prompts the user for confirmation", func() {
			err := destroy.Execute([]string{}, storage.State{
				BOSH: storage.BOSH{
					DirectorName: "some-director",
				},
				EnvID: "some-lake",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.PromptCall.Receives.Message).To(Equal(`Are you sure you want to delete infrastructure for "some-lake"? This operation cannot be undone!`))
			Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(1))
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not delete anything", func() {
				err := destroy.Execute([]string{}, storage.State{
					BOSH: storage.BOSH{
						DirectorName: "some-director",
					},
					EnvID: "some-lake",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal(`Are you sure you want to delete infrastructure for "some-lake"? This operation cannot be undone!`))
				Expect(logger.StepCall.Receives.Message).To(Equal("exiting"))
				Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(0))
			})
		})

		It("invokes bosh delete", func() {
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
				state := storage.State{
					EnvID: "unintialized",
					LB:    storage.LB{Type: "lb-type", Domain: "lb-domain"},
				}
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(plan.IsInitializedCall.CallCount).To(Equal(1))
				Expect(plan.IsInitializedCall.Receives.State).To(Equal(state))
				Expect(plan.InitializePlanCall.CallCount).To(Equal(1))
				Expect(plan.InitializePlanCall.Receives.State).To(Equal(state))
				Expect(plan.InitializePlanCall.Receives.Plan).To(Equal(commands.PlanConfig{
					Name: "unintialized",
					LB:   storage.LB{Type: "lb-type", Domain: "lb-domain"},
				}))
			})
		})

		Context("when the environment is not paved", func() {
			It("deletes the directory without attempting to destroy bosh or terraform", func() {
				terraformManager.IsPavedCall.Returns.IsPaved = false

				err := destroy.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(0))
				Expect(boshManager.DeleteJumpboxCall.CallCount).To(Equal(0))
				Expect(terraformManager.DestroyCall.CallCount).To(Equal(0))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{}))
			})
		})

		Context("failure cases", func() {
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
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
					IAAS: "aws",
					BOSH: storage.BOSH{State: map[string]interface{}{"key": "value"}},
				}
			})

			It("calls terraform destroy and deletes the state file", func() {
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				expectedState := state
				expectedState.BOSH = storage.BOSH{}
				Expect(terraformManager.SetupCall.Receives.BBLState).To(Equal(expectedState))
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
				})

				It("saves the partially destroyed tf state", func() {
					err := destroy.Execute([]string{}, state)
					Expect(err).To(MatchError("failed to destroy"))

					Expect(terraformManager.SetupCall.CallCount).To(Equal(1))
					Expect(terraformManager.SetupCall.Receives.BBLState).To(Equal(expectedBBLState))
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

			Context("reentrance", func() {
				Context("when NoDirector is true", func() {
					It("does not attempt to delete the bosh director", func() {
						state.NoDirector = true
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintlnCall.Receives.Message).To(Equal("No BOSH director, skipping..."))
						Expect(logger.StepCall.Messages).NotTo(ContainElement("destroying bosh director"))
						Expect(boshManager.DeleteDirectorCall.CallCount).To(Equal(0))
					})
				})

			})
		})

		Context("failure cases", func() {
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
							BOSH: storage.BOSH{State: map[string]interface{}{"error": "state"}},
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
