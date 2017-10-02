package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	commandsFakes "github.com/cloudfoundry/bosh-bootloader/commands/fakes"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteLBs", func() {
	var (
		command commands.DeleteLBs

		stateValidator       *fakes.StateValidator
		logger               *fakes.Logger
		boshManager          *fakes.BOSHManager
		environmentValidator *fakes.EnvironmentValidator
		cloudConfigManager   *fakes.CloudConfigManager
		terraformManager     *commandsFakes.TerraformApplier
		stateStore           *fakes.StateStore

		incomingState storage.State
	)

	BeforeEach(func() {
		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"
		environmentValidator = &fakes.EnvironmentValidator{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		terraformManager = &commandsFakes.TerraformApplier{}
		stateStore = &fakes.StateStore{}

		incomingState = storage.State{
			LB: storage.LB{
				Type: "concourse",
				Cert: "some-cert",
				Key:  "some-key",
			},
			TFState: "some-tf-state",
		}

		command = commands.NewDeleteLBs(logger, stateValidator, boshManager, cloudConfigManager, stateStore, environmentValidator, terraformManager)
	})

	Describe("CheckFastFails", func() {
		Context("when state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			})

			It("returns an error", func() {
				err := command.CheckFastFails([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})
		})

		Context("when the BOSH version is less than 2.0.24 and there is a director", func() {
			It("returns a helpful error message", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{}, storage.State{
					IAAS: "aws",
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})
		})

		Context("when the BOSH version is less than 2.0.24 and there is no director", func() {
			It("does not fast fail", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{}, storage.State{
					IAAS:       "gcp",
					NoDirector: true,
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Execute", func() {
		It("deletes the load balancers", func() {
			err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			By("validating the environment", func() {
				Expect(environmentValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(environmentValidator.ValidateCall.Receives.State).To(Equal(incomingState))
			})

			By("updating cloud config", func() {
				Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Type).To(BeEmpty())
				Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Cert).To(BeEmpty())
				Expect(cloudConfigManager.UpdateCall.Receives.State.LB.Key).To(BeEmpty())
			})

			By("running terraform apply to delete lbs and certificate", func() {
				Expect(terraformManager.ApplyCallCount()).To(Equal(1))

				expectedTerraformState := incomingState
				expectedTerraformState.LB = storage.LB{}

				Expect(terraformManager.ApplyArgsForCall(0)).To(Equal(expectedTerraformState))
			})

			By("saving state with no lb type", func() {
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
					LB: storage.LB{},
				}))
			})
		})

		Context("when the bbl env was created without a bosh director", func() {
			It("does not try to update the cloud config", func() {
				state := storage.State{
					NoDirector: true,
					LB: storage.LB{
						Type: "concourse",
					},
				}
				terraformManager.ApplyReturns(state, nil)

				err := command.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("when there is no lb", func() {
			It("returns an error", func() {
				err := command.Execute([]string{}, storage.State{
					TFState: "some-tf-state",
				})
				Expect(err).To(MatchError(commands.LBNotFound))
			})
		})

		Context("when an error occurs", func() {
			Context("when the environment validator fails", func() {
				It("returns an error", func() {
					environmentValidator.ValidateCall.Returns.Error = errors.New("mesclun")

					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Environment validate: mesclun"))
				})
			})
			Context("when terraform manager fails to apply the second time with terraformManagerError", func() {
				It("return an error", func() {
					terraformManager.ApplyReturns(storage.State{}, errors.New("apply failed"))

					err := command.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("apply failed"))
				})
			})

			Context("when terraform manager fails to apply with non-terraformManagerError", func() {
				var (
					managerError *fakes.TerraformManagerError
				)

				BeforeEach(func() {
					managerError = &fakes.TerraformManagerError{}
					managerError.BBLStateCall.Returns.BBLState = storage.State{
						TFState: "some-partial-tf-state",
					}
					managerError.ErrorCall.Returns = "cannot apply"

					terraformManager.ApplyReturns(storage.State{}, managerError)
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("cannot apply"))

					Expect(stateStore.SetCall.CallCount).To(Equal(1))
					Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
						TFState: "some-partial-tf-state",
					}))
				})

				Context("when the terraform manager error fails to return a bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.Error = errors.New("failed to retrieve bbl state")
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute([]string{}, incomingState)
						Expect(err).To(MatchError("the following errors occurred:\ncannot apply,\nfailed to retrieve bbl state"))
					})
				})
			})

			Context("when cloud config manager fails to update", func() {
				BeforeEach(func() {
					cloudConfigManager.UpdateCall.Returns.Error = errors.New("spinach")
				})

				It("return an error", func() {
					err := command.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("Update cloud config: spinach"))
				})
			})

			Context("when the state fails to save lb type", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("kale")}}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("Save state after delete lbs: kale"))
				})
			})
		})

		Context("when --skip-if-missing is provided", func() {
			DescribeTable("no-ops", func(state storage.State) {
				err := command.Execute([]string{
					"--skip-if-missing",
				}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCallCount()).To(Equal(0))
				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`no lb type exists, skipping...`))
			},
				Entry("no-ops when LB type does not exist in state LB", storage.State{
					LB: storage.LB{
						Type: "",
					},
				}),
			)
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -unknown-flag"))

					Expect(terraformManager.ApplyCallCount()).To(Equal(0))
				})
			})
		})
	})
})
