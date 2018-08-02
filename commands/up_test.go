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

var _ = Describe("Up", func() {
	var (
		command commands.Up

		plan                 *fakes.Plan
		boshManager          *fakes.BOSHManager
		terraformManager     *fakes.TerraformManager
		cloudConfigManager   *fakes.CloudConfigManager
		runtimeConfigManager *fakes.RuntimeConfigManager
		stateStore           *fakes.StateStore
	)

	BeforeEach(func() {
		plan = &fakes.Plan{}
		boshManager = &fakes.BOSHManager{}
		terraformManager = &fakes.TerraformManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		runtimeConfigManager = &fakes.RuntimeConfigManager{}
		stateStore = &fakes.StateStore{}

		command = commands.NewUp(plan, boshManager, cloudConfigManager, runtimeConfigManager, stateStore, terraformManager)
	})

	Describe("CheckFastFails", func() {
		It("returns CheckFastFails on Plan", func() {
			plan.CheckFastFailsCall.Returns.Error = errors.New("banana")
			err := command.CheckFastFails([]string{}, storage.State{Version: 999})

			Expect(err).To(MatchError("banana"))
			Expect(plan.CheckFastFailsCall.Receives.SubcommandFlags).To(Equal([]string{}))
			Expect(plan.CheckFastFailsCall.Receives.State).To(Equal(storage.State{Version: 999}))
		})
	})

	Describe("Execute", func() {
		var (
			incomingState       storage.State
			planState           storage.State
			planConfig          commands.PlanConfig
			terraformApplyState storage.State
			createJumpboxState  storage.State
			createDirectorState storage.State
			terraformOutputs    terraform.Outputs
		)
		BeforeEach(func() {
			planConfig = commands.PlanConfig{Name: "some-name"}
			plan.ParseArgsCall.Returns.Config = planConfig

			incomingState = storage.State{LatestTFOutput: "incoming-state", IAAS: "some-iaas"}

			planState = storage.State{LatestTFOutput: "plan-state", IAAS: "some-iaas"}
			plan.InitializePlanCall.Returns.State = planState

			terraformApplyState = storage.State{LatestTFOutput: "terraform-apply-call", IAAS: "some-iaas"}
			terraformManager.ApplyCall.Returns.BBLState = terraformApplyState

			createJumpboxState = storage.State{LatestTFOutput: "create-jumpbox-call", IAAS: "some-iaas"}
			boshManager.CreateJumpboxCall.Returns.State = createJumpboxState

			createDirectorState = storage.State{LatestTFOutput: "create-director-call", IAAS: "some-iaas"}
			boshManager.CreateDirectorCall.Returns.State = createDirectorState

			terraformOutputs = terraform.Outputs{
				Map: map[string]interface{}{
					"jumpbox_url": "some-jumpbox-url",
				},
			}
			terraformManager.GetOutputsCall.Returns.Outputs = terraformOutputs

			plan.IsInitializedCall.Returns.IsInitialized = true
		})

		Context("when bbl plan has been run", func() {
			It("applies without re-initializing", func() {
				err := command.Execute([]string{"some", "flags"}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(plan.IsInitializedCall.CallCount).To(Equal(1))
				Expect(plan.IsInitializedCall.Receives.State).To(Equal(incomingState))

				Expect(plan.ParseArgsCall.CallCount).To(Equal(1))
				Expect(plan.ParseArgsCall.Receives.Args).To(Equal([]string{"some", "flags"}))
				Expect(plan.ParseArgsCall.Receives.State).To(Equal(incomingState))

				Expect(plan.InitializePlanCall.CallCount).To(Equal(0))

				Expect(terraformManager.SetupCall.CallCount).To(Equal(0))

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(incomingState))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(terraformApplyState))

				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))

				Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(0))
				Expect(boshManager.CreateJumpboxCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateJumpboxCall.Receives.State).To(Equal(terraformApplyState))
				Expect(boshManager.CreateJumpboxCall.Receives.TerraformOutputs).To(Equal(terraformOutputs))
				Expect(stateStore.SetCall.Receives[1].State).To(Equal(createJumpboxState))

				Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(0))
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(createJumpboxState))
				Expect(boshManager.CreateDirectorCall.Receives.TerraformOutputs).To(Equal(terraformOutputs))
				Expect(stateStore.SetCall.Receives[2].State).To(Equal(createDirectorState))

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
				Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(createDirectorState))

				Expect(runtimeConfigManager.UpdateCall.CallCount).To(Equal(1))
				Expect(runtimeConfigManager.UpdateCall.Receives.State).To(Equal(createDirectorState))

				Expect(stateStore.SetCall.CallCount).To(Equal(3))
			})
		})

		Context("if parse args fails", func() {
			It("returns an error if parse args fails", func() {
				plan.ParseArgsCall.Returns.Error = errors.New("canteloupe")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("canteloupe"))
			})
		})

		Context("when nothing is initialized", func() {
			BeforeEach(func() {
				plan.IsInitializedCall.Returns.IsInitialized = false
			})
			It("calls bbl plan", func() {
				err := command.Execute([]string{"some", "flags"}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(plan.IsInitializedCall.CallCount).To(Equal(1))
				Expect(plan.IsInitializedCall.Receives.State).To(Equal(incomingState))

				Expect(plan.InitializePlanCall.CallCount).To(Equal(1))
				Expect(plan.InitializePlanCall.Receives.Plan).To(Equal(planConfig))
				Expect(plan.InitializePlanCall.Receives.State).To(Equal(incomingState))

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(planState))
			})
		})

		Describe("failure cases", func() {
			Context("when parse args fails", func() {
				BeforeEach(func() {
					plan.ParseArgsCall.Returns.Error = errors.New("apple")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("apple"))
				})
			})

			Context("when the terraform manager fails with non terraformManagerError", func() {
				BeforeEach(func() {
					terraformManager.ApplyCall.Returns.Error = errors.New("passionfruit")
				})

				It("returns the error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("passionfruit"))
				})
			})

			Context("when saving the state fails after terraform apply", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("kiwi")}}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Save state after terraform apply: kiwi"))
				})
			})

			Context("when the terraform manager cannot get terraform outputs", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("raspberry")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Parse terraform outputs: raspberry"))
				})
			})

			Context("when the jumpbox cannot be deployed", func() {
				BeforeEach(func() {
					boshManager.CreateJumpboxCall.Returns.Error = errors.New("pineapple")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Create jumpbox: pineapple"))
				})
			})

			Context("when saving the state fails after create jumpbox", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {Error: errors.New("kiwi")}}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Save state after create jumpbox: kiwi"))
				})
			})

			Context("when bosh cannot be deployed", func() {
				BeforeEach(func() {
					boshManager.CreateDirectorCall.Returns.Error = errors.New("pineapple")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Create bosh director: pineapple"))
				})
			})

			Context("when saving the state fails after create director", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {Error: errors.New("kiwi")}}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Save state after create director: kiwi"))
				})
			})

			Context("when the cloud config cannot be uploaded", func() {
				BeforeEach(func() {
					cloudConfigManager.UpdateCall.Returns.Error = errors.New("coconut")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Update cloud config: coconut"))
				})
			})

			Context("when the runtime config cannot be uploaded", func() {
				BeforeEach(func() {
					runtimeConfigManager.UpdateCall.Returns.Error = errors.New("grapefruit")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Update runtime config: grapefruit"))
				})
			})

			Context("when terraform manager apply fails", func() {
				var partialState storage.State

				BeforeEach(func() {
					partialState = storage.State{
						LatestTFOutput: "some terraform error",
					}
					terraformManager.ApplyCall.Returns.BBLState = partialState
					terraformManager.ApplyCall.Returns.Error = errors.New("grapefruit")
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

			Context("when the bosh manager fails to create a jumpbox with ManagerCreateError", func() {
				var partialState storage.State

				BeforeEach(func() {
					partialState = storage.State{LatestTFOutput: "some terraform error"}
					expectedError := bosh.NewManagerCreateError(partialState, errors.New("rambutan"))
					boshManager.CreateJumpboxCall.Returns.Error = expectedError
				})

				It("returns the error and saves the state", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Create jumpbox: rambutan"))

					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State).To(Equal(partialState))
				})

				Context("when it fails to save the state", func() {
					BeforeEach(func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("lychee")}}
					})

					It("returns a compound error", func() {
						err := command.Execute([]string{}, storage.State{})
						Expect(err).To(MatchError("Save state after jumpbox create error: rambutan, lychee"))

						Expect(stateStore.SetCall.CallCount).To(Equal(2))
						Expect(stateStore.SetCall.Receives[1].State).To(Equal(partialState))
					})
				})
			})

			Context("when the bosh manager fails to create the director with ManagerCreateError", func() {
				var partialState storage.State

				BeforeEach(func() {
					partialState = storage.State{LatestTFOutput: "some terraform error"}
					expectedError := bosh.NewManagerCreateError(partialState, errors.New("rambutan"))
					boshManager.CreateDirectorCall.Returns.Error = expectedError
				})

				It("returns the error and saves the state", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Create bosh director: rambutan"))

					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State).To(Equal(partialState))
				})

				Context("when it fails to save the state", func() {
					BeforeEach(func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("lychee")}}
					})

					It("returns a compound error", func() {
						err := command.Execute([]string{}, storage.State{})
						Expect(err).To(MatchError("Save state after bosh director create error: rambutan, lychee"))

						Expect(stateStore.SetCall.CallCount).To(Equal(3))
						Expect(stateStore.SetCall.Receives[2].State).To(Equal(partialState))
					})
				})
			})
		})
	})

	Describe("ParseArgs", func() {
		It("returns ParseArgs on Plan", func() {
			plan.ParseArgsCall.Returns.Config = commands.PlanConfig{Name: "environment name"}
			config, err := command.ParseArgs([]string{"--name", "environment name"}, storage.State{ID: "some-state-id"})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.ParseArgsCall.Receives.Args).To(Equal([]string{"--name", "environment name"}))
			Expect(plan.ParseArgsCall.Receives.State).To(Equal(storage.State{ID: "some-state-id"}))
			Expect(config.Name).To(Equal("environment name"))
		})
	})
})
