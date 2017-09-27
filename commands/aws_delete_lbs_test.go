package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	commandsFakes "github.com/cloudfoundry/bosh-bootloader/commands/fakes"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete LBs", func() {
	var (
		command              commands.AWSDeleteLBs
		environmentValidator *fakes.EnvironmentValidator
		cloudConfigManager   *fakes.CloudConfigManager
		terraformManager     *commandsFakes.TerraformApplier
		stateStore           *fakes.StateStore

		incomingState storage.State
	)

	BeforeEach(func() {
		environmentValidator = &fakes.EnvironmentValidator{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		terraformManager = &commandsFakes.TerraformApplier{}
		stateStore = &fakes.StateStore{}

		incomingState = storage.State{
			AWS: storage.AWS{
				Region: "some-region",
			},
			KeyPair: storage.KeyPair{
				Name: "some-keypair",
			},
			EnvID: "some-env-id",
			LB: storage.LB{
				Type: "concourse",
				Cert: "some-cert",
				Key:  "some-key",
			},
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
			TFState: "some-tf-state",
		}

		command = commands.NewAWSDeleteLBs(cloudConfigManager, stateStore, environmentValidator, terraformManager)
	})

	Describe("Execute", func() {
		It("deletes the load balancers", func() {
			err := command.Execute(incomingState)
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

				err := command.Execute(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("when there is no lb", func() {
			It("returns an error", func() {
				err := command.Execute(storage.State{
					TFState: "some-tf-state",
				})
				Expect(err).To(MatchError(commands.LBNotFound))
			})
		})

		Context("when an error occurs", func() {
			Context("when the environment validator fails", func() {
				It("returns an error", func() {
					environmentValidator.ValidateCall.Returns.Error = errors.New("validate failed")

					err := command.Execute(incomingState)
					Expect(err).To(MatchError("validate failed"))
				})
			})
			Context("when terraform manager fails to apply the second time with terraformManagerError", func() {
				It("return an error", func() {
					terraformManager.ApplyReturns(storage.State{}, errors.New("apply failed"))

					err := command.Execute(incomingState)
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

				It("return an error", func() {
					err := command.Execute(incomingState)
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
						err := command.Execute(incomingState)
						Expect(err).To(MatchError("the following errors occurred:\ncannot apply,\nfailed to retrieve bbl state"))
					})
				})
			})

			It("return an error when cloud config manager fails to update", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("update failed")
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("update failed"))
			})

			It("returns an error when the state fails to save lb type", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to save state")}}
				err := command.Execute(incomingState)
				Expect(err).To(MatchError("failed to save state"))
			})
		})
	})
})
