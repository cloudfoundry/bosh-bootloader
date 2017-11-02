package commands_test

import (
	"errors"
	"io/ioutil"

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

		plan               *fakes.Plan
		boshManager        *fakes.BOSHManager
		terraformManager   *fakes.TerraformManager
		cloudConfigManager *fakes.CloudConfigManager
		stateStore         *fakes.StateStore
		envIDManager       *fakes.EnvIDManager

		tempDir string
	)

	BeforeEach(func() {
		plan = &fakes.Plan{}

		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"

		terraformManager = &fakes.TerraformManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		stateStore = &fakes.StateStore{}
		envIDManager = &fakes.EnvIDManager{}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		stateStore.GetBblDirCall.Returns.Directory = tempDir

		command = commands.NewUp(plan, boshManager, cloudConfigManager, stateStore, envIDManager, terraformManager)
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
			iaasState           storage.State
			envIDManagerState   storage.State
			terraformApplyState storage.State
			createJumpboxState  storage.State
			createDirectorState storage.State
			terraformOutputs    terraform.Outputs
		)
		BeforeEach(func() {
			plan.ParseArgsCall.Returns.Config = commands.UpConfig{Name: "some-name"}

			incomingState = storage.State{TFState: "incoming-state", IAAS: "some-iaas"}
			iaasState = storage.State{TFState: "iaas-state", IAAS: "some-iaas"}

			envIDManagerState = storage.State{TFState: "env-id-sync-call", IAAS: "some-iaas"}
			envIDManager.SyncCall.Returns.State = envIDManagerState

			terraformApplyState = storage.State{TFState: "terraform-apply-call", IAAS: "some-iaas"}
			terraformManager.ApplyCall.Returns.BBLState = terraformApplyState

			createJumpboxState = storage.State{TFState: "create-jumpbox-call", IAAS: "some-iaas"}
			boshManager.CreateJumpboxCall.Returns.State = createJumpboxState

			createDirectorState = storage.State{TFState: "create-director-call", IAAS: "some-iaas"}
			boshManager.CreateDirectorCall.Returns.State = createDirectorState

			terraformOutputs = terraform.Outputs{
				Map: map[string]interface{}{
					"jumpbox_url": "some-jumpbox-url",
				},
			}
			terraformManager.GetOutputsCall.Returns.Outputs = terraformOutputs

			terraformManager.IsInitializedCall.Returns.IsInitialized = true
			boshManager.IsJumpboxInitializedCall.Returns.IsInitialized = true
			boshManager.IsDirectorInitializedCall.Returns.IsInitialized = true
		})

		Context("when bbl plan has been run", func() {
			It("applies without re-initializing", func() {
				err := command.Execute([]string{"some", "flags"}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(plan.ParseArgsCall.CallCount).To(Equal(1))
				Expect(plan.ParseArgsCall.Receives.Args).To(Equal([]string{"some", "flags"}))
				Expect(plan.ParseArgsCall.Receives.State).To(Equal(incomingState))

				Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
				Expect(envIDManager.SyncCall.Receives.State).To(Equal(incomingState))
				Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-name"))
				Expect(stateStore.SetCall.Receives[0].State).To(Equal(envIDManagerState))

				Expect(terraformManager.IsInitializedCall.CallCount).To(Equal(1))
				Expect(terraformManager.InitCall.CallCount).To(Equal(0))

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(envIDManagerState))
				Expect(stateStore.SetCall.Receives[1].State).To(Equal(terraformApplyState))

				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(terraformApplyState))

				Expect(boshManager.IsJumpboxInitializedCall.CallCount).To(Equal(1))
				Expect(boshManager.IsJumpboxInitializedCall.Receives.IAAS).To(Equal("some-iaas"))
				Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(0))
				Expect(boshManager.CreateJumpboxCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateJumpboxCall.Receives.State).To(Equal(terraformApplyState))
				Expect(boshManager.CreateJumpboxCall.Receives.TerraformOutputs).To(Equal(terraformOutputs))
				Expect(stateStore.SetCall.Receives[2].State).To(Equal(createJumpboxState))

				Expect(boshManager.IsDirectorInitializedCall.CallCount).To(Equal(1))
				Expect(boshManager.IsDirectorInitializedCall.Receives.IAAS).To(Equal("some-iaas"))
				Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(0))
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(1))
				Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(createJumpboxState))
				Expect(boshManager.CreateDirectorCall.Receives.TerraformOutputs).To(Equal(terraformOutputs))
				Expect(stateStore.SetCall.Receives[3].State).To(Equal(createDirectorState))

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
				Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(createDirectorState))

				Expect(stateStore.SetCall.CallCount).To(Equal(4))
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
				terraformManager.IsInitializedCall.Returns.IsInitialized = false
				boshManager.IsJumpboxInitializedCall.Returns.IsInitialized = false
				boshManager.IsDirectorInitializedCall.Returns.IsInitialized = false
			})
			It("calls bbl plan", func() {
				err := command.Execute([]string{"some", "flags"}, incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.ExecuteCall.CallCount).To(Equal(1))
				Expect(plan.ExecuteCall.Receives.Args).To(Equal([]string{"some", "flags"}))
				Expect(plan.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when terraform is not initialized yet", func() {
			BeforeEach(func() {
				terraformManager.IsInitializedCall.Returns.IsInitialized = false
			})
			It("calls bbl plan", func() {
				err := command.Execute([]string{"some", "flags"}, incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.ExecuteCall.CallCount).To(Equal(1))
				Expect(plan.ExecuteCall.Receives.Args).To(Equal([]string{"some", "flags"}))
				Expect(plan.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when the jumpbox is not initialized yet", func() {
			BeforeEach(func() {
				boshManager.IsJumpboxInitializedCall.Returns.IsInitialized = false
			})
			It("calls init on the manager", func() {
				err := command.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.ExecuteCall.CallCount).To(Equal(1))
				Expect(plan.ExecuteCall.Receives.Args).To(Equal([]string{}))
				Expect(plan.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when the director is not initialized yet", func() {
			BeforeEach(func() {
				boshManager.IsDirectorInitializedCall.Returns.IsInitialized = false
			})
			It("calls init on the manager", func() {
				err := command.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.ExecuteCall.CallCount).To(Equal(1))
				Expect(plan.ExecuteCall.Receives.Args).To(Equal([]string{}))
				Expect(plan.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when --no-director flag is passed", func() {
			BeforeEach(func() {
				plan.ParseArgsCall.Returns.Config = commands.UpConfig{NoDirector: true}
			})

			It("sets NoDirector to true on the state", func() {
				err := command.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDManager.SyncCall.Receives.State.NoDirector).To(BeTrue())
			})
		})

		Context("when the config or state has the no-director flag set", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true
			})

			It("does not create a bosh or cloud config", func() {
				err := command.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(0))
				Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(0))
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[1].State.NoDirector).To(BeTrue())
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
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

			Context("when the config has the no-director flag set and the bbl state has a bosh director", func() {
				BeforeEach(func() {
					incomingState = storage.State{BOSH: storage.BOSH{DirectorName: "some-director"}}
					plan.ParseArgsCall.Returns.Config = commands.UpConfig{NoDirector: true}
				})

				It("fast fails", func() {
					err := command.Execute([]string{}, incomingState)
					Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
				})
			})

			Context("when the env id manager fails", func() {
				BeforeEach(func() {
					envIDManager.SyncCall.Returns.Error = errors.New("apple")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Env id manager sync: apple"))
				})
			})

			Context("when saving the state fails after env id sync", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("kiwi")}}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Save state after sync: kiwi"))
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
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {Error: errors.New("kiwi")}}
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
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {Error: errors.New("kiwi")}}
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
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {Error: errors.New("kiwi")}}
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

			Context("when the terraform manager fails with terraformManagerError", func() {
				var (
					managerError *fakes.TerraformManagerError
					partialState storage.State
				)

				BeforeEach(func() {
					managerError = &fakes.TerraformManagerError{}
					partialState = storage.State{
						TFState: "some-partial-tf-state",
					}
					managerError.BBLStateCall.Returns.BBLState = partialState
					managerError.ErrorCall.Returns = "grapefruit"
					terraformManager.ApplyCall.Returns.Error = managerError
				})

				It("saves the bbl state and returns the error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("grapefruit"))

					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State).To(Equal(partialState))
				})

				Context("when the applier fails and we cannot retrieve the updated bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.Error = errors.New("failed to retrieve bbl state")
					})

					It("returns an error", func() {
						err := command.Execute([]string{}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\ngrapefruit,\nfailed to retrieve bbl state"))
					})
				})

				Context("when we fail to set the bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.BBLState = partialState
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set bbl state")}}
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute([]string{}, storage.State{})
						Expect(err).To(MatchError("the following errors occurred:\ngrapefruit,\nfailed to set bbl state"))
					})
				})
			})

			Context("when the bosh manager fails with BOSHManagerCreate error", func() {
				var partialState storage.State

				BeforeEach(func() {
					partialState = storage.State{TFState: "some-partial-tf-state"}
					expectedError := bosh.NewManagerCreateError(partialState, errors.New("rambutan"))
					boshManager.CreateDirectorCall.Returns.Error = expectedError
				})

				It("returns the error and saves the state", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Create bosh director: rambutan"))

					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives[3].State).To(Equal(partialState))
				})

				Context("when it fails to save the state", func() {
					BeforeEach(func() {
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("lychee")}}
					})

					It("returns a compound error", func() {
						err := command.Execute([]string{}, storage.State{})
						Expect(err).To(MatchError("Save state after bosh director create error: rambutan, lychee"))

						Expect(stateStore.SetCall.CallCount).To(Equal(4))
						Expect(stateStore.SetCall.Receives[3].State).To(Equal(partialState))
					})
				})
			})
		})
	})

	Describe("ParseArgs", func() {
		It("returns ParseArgs on Plan", func() {
			plan.ParseArgsCall.Returns.Config = commands.UpConfig{OpsFile: "some-path"}
			config, err := command.ParseArgs([]string{"--ops-file", "some-path"}, storage.State{ID: "some-state-id"})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.ParseArgsCall.Receives.Args).To(Equal([]string{"--ops-file", "some-path"}))
			Expect(plan.ParseArgsCall.Receives.State).To(Equal(storage.State{ID: "some-state-id"}))
			Expect(config.OpsFile).To(Equal("some-path"))
		})
	})
})
