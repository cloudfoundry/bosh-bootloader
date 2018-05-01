package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validate", func() {
	var (
		command          commands.Validate
		plan             *fakes.Plan
		stateStore       *fakes.StateStore
		terraformManager *fakes.TerraformManager
	)

	BeforeEach(func() {
		plan = &fakes.Plan{}
		stateStore = &fakes.StateStore{}
		terraformManager = &fakes.TerraformManager{}

		command = commands.NewValidate(plan, stateStore, terraformManager)
	})

	Describe("CheckFastFails", func() {
		It("returns CheckFastFails on Plan", func() {
			plan.CheckFastFailsCall.Returns.Error = errors.New("banana")
			err := command.CheckFastFails([]string{}, storage.State{Version: 999, IAAS: "some-iaas"})

			Expect(err).To(MatchError("banana"))
			Expect(plan.CheckFastFailsCall.Receives.SubcommandFlags).To(Equal([]string{}))
			Expect(plan.CheckFastFailsCall.Receives.State).To(Equal(storage.State{Version: 999, IAAS: "some-iaas"}))
		})

		Context("without an iaas", func() {
			It("returns an error", func() {
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})
				Expect(err).To(MatchError("bbl state has not been initialized yet, please run bbl plan"))

				Expect(plan.IsInitializedCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("Execute", func() {
		var (
			incomingState storage.State
			expectedState storage.State
		)

		BeforeEach(func() {
			incomingState = storage.State{LatestTFOutput: "not validated yet", IAAS: "some-iaas"}
			expectedState = storage.State{LatestTFOutput: "validated", IAAS: "some-iaas"}

			terraformManager.ValidateCall.Returns.BBLState = expectedState

			plan.IsInitializedCall.Returns.IsInitialized = true
		})

		It("validates terraform", func() {
			err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.IsInitializedCall.CallCount).To(Equal(1))
			Expect(plan.IsInitializedCall.Receives.State).To(Equal(incomingState))

			Expect(terraformManager.InitCall.CallCount).To(Equal(1))
			Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(incomingState))

			Expect(terraformManager.ValidateCall.CallCount).To(Equal(1))
			Expect(terraformManager.ValidateCall.Receives.BBLState).To(Equal(incomingState))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(expectedState))
		})

		Describe("failure cases", func() {
			Context("when plan hasn't been initialized", func() {
				BeforeEach(func() {
					plan.IsInitializedCall.Returns.IsInitialized = false
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("bbl state has not been initialized yet, please run bbl plan"))
				})
			})

			Context("when terraform manager fails to run terraform init", func() {
				BeforeEach(func() {
					terraformManager.InitCall.Returns.Error = errors.New("passionfruit")
				})

				It("returns the error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("passionfruit"))
				})
			})

			Context("when the terraform manager fails with non terraformManagerError", func() {
				BeforeEach(func() {
					terraformManager.ValidateCall.Returns.Error = errors.New("passionfruit")
				})

				It("returns the error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("passionfruit"))
				})
			})

			Context("when saving the state fails after terraform validate", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("kiwi")}}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Save state after terraform validate: kiwi"))
				})
			})

			Context("when terraform manager validate fails", func() {
				var partialState storage.State

				BeforeEach(func() {
					partialState = storage.State{
						LatestTFOutput: "some terraform error",
					}
					terraformManager.ValidateCall.Returns.BBLState = partialState
					terraformManager.ValidateCall.Returns.Error = errors.New("grapefruit")
				})

				It("saves the bbl state and returns the error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("grapefruit"))

					Expect(stateStore.SetCall.CallCount).To(Equal(1))
					Expect(stateStore.SetCall.Receives[0].State).To(Equal(partialState))
				})

				Context("when we fail to set the bbl state", func() {
					BeforeEach(func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set bbl state")}}
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute([]string{}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\ngrapefruit,\nfailed to set bbl state"))
					})
				})
			})
		})
	})
})
