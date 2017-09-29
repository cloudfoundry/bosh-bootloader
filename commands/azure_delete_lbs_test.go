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
		command            commands.AzureDeleteLBs
		cloudConfigManager *fakes.CloudConfigManager
		terraformManager   *commandsFakes.TerraformApplier
		stateStore         *fakes.StateStore

		incomingState storage.State
	)

	BeforeEach(func() {
		cloudConfigManager = &fakes.CloudConfigManager{}
		terraformManager = &commandsFakes.TerraformApplier{}
		stateStore = &fakes.StateStore{}

		incomingState = storage.State{
			EnvID: "some-env-id",
			LB: storage.LB{
				Type: "concourse",
				Cert: "some-cert",
				Key:  "some-key",
			},
			TFState: "some-tf-state",
		}

		command = commands.NewAzureDeleteLBs(cloudConfigManager,
			stateStore, terraformManager)
	})

	Describe("Execute", func() {
		It("deletes the load balancers", func() {
			err := command.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())

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
				}

				terraformManager.ApplyReturns(state, nil)

				err := command.Execute(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("when an error occurs", func() {
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

			Context("when cloud config manager fails to update", func() {
				BeforeEach(func() {
					cloudConfigManager.UpdateCall.Returns.Error = errors.New("update failed")
				})

				It("returns an error", func() {
					err := command.Execute(incomingState)
					Expect(err).To(MatchError("update failed"))
				})
			})

			Context("when the state fails to save lb type", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to save state")}}
				})
				It("returns an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to save state")}}
					err := command.Execute(incomingState)
					Expect(err).To(MatchError("failed to save state"))
				})
			})
		})
	})
})
