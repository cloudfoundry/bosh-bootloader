package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSUp", func() {
	Describe("Execute", func() {
		var (
			command            commands.AWSUp
			boshManager        *fakes.BOSHManager
			terraformManager   *fakes.TerraformManager
			cloudConfigManager *fakes.CloudConfigManager
			stateStore         *fakes.StateStore
			envIDManager       *fakes.EnvIDManager

			upConfig      commands.UpConfig
			incomingState storage.State

			envIDManagerState   storage.State
			terraformApplyState storage.State
			createJumpboxState  storage.State
			createDirectorState storage.State
		)

		BeforeEach(func() {
			upConfig = commands.UpConfig{Name: "some-name"}
			incomingState = storage.State{TFState: "incoming-state"}

			envIDManager = &fakes.EnvIDManager{}
			envIDManagerState = storage.State{TFState: "env-id-sync-call"}
			envIDManager.SyncCall.Returns.State = envIDManagerState

			terraformManager = &fakes.TerraformManager{}
			terraformApplyState = storage.State{TFState: "terraform-apply-call"}
			terraformManager.ApplyCall.Returns.BBLState = terraformApplyState

			boshManager = &fakes.BOSHManager{}
			createJumpboxState = storage.State{TFState: "create-jumpbox-call"}
			boshManager.CreateJumpboxCall.Returns.State = createJumpboxState

			createDirectorState = storage.State{TFState: "create-director-call"}
			boshManager.CreateDirectorCall.Returns.State = createDirectorState

			cloudConfigManager = &fakes.CloudConfigManager{}
			stateStore = &fakes.StateStore{}

			command = commands.NewAWSUp(boshManager,
				cloudConfigManager, stateStore,
				envIDManager, terraformManager)
		})

		It("creates infrastructure", func() {
			err := command.Execute(upConfig, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.ValidateVersionCall.CallCount).To(Equal(1))

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(envIDManager.SyncCall.Receives.State).To(Equal(incomingState))
			Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-name"))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(envIDManagerState))

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(envIDManagerState))
			Expect(stateStore.SetCall.Receives[1].State).To(Equal(terraformApplyState))

			Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(terraformApplyState))

			Expect(boshManager.CreateJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.CreateJumpboxCall.Receives.State).To(Equal(terraformApplyState))
			Expect(stateStore.SetCall.Receives[2].State).To(Equal(createJumpboxState))

			Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(createJumpboxState))
			Expect(stateStore.SetCall.Receives[3].State).To(Equal(createDirectorState))

			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(createDirectorState))

			Expect(stateStore.SetCall.CallCount).To(Equal(4))
		})

		Context("when the config has ops files", func() {
			BeforeEach(func() {
				opsFile, err := ioutil.TempFile("", "ops-file")
				Expect(err).NotTo(HaveOccurred())

				opsFilePath := opsFile.Name()
				opsFileContents := "some-ops-file-contents"
				err = ioutil.WriteFile(opsFilePath, []byte(opsFileContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
				upConfig = commands.UpConfig{OpsFile: opsFilePath}
			})

			It("passes the ops file contents to the bosh manager", func() {
				err := command.Execute(upConfig, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateDirectorCall.Receives.State.BOSH.UserOpsFile).To(Equal("some-ops-file-contents"))
			})
		})

		Context("when the config or state has the no-director flag set", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true
			})

			It("does not create a bosh or cloud config", func() {
				err := command.Execute(upConfig, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.Receives[1].State.NoDirector).To(BeTrue())
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
			})

		})

		Describe("failure cases", func() {
			It("returns an error if terraform manager version validator fails", func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("grape")
				err := command.Execute(commands.UpConfig{}, storage.State{})

				Expect(err).To(MatchError("Terraform validate version: grape"))
			})

			It("fast fails when the config has the no-director flag set and the bbl state has a bosh director", func() {
				upConfig = commands.UpConfig{NoDirector: true}
				incomingState = storage.State{BOSH: storage.BOSH{DirectorName: "some-director"}}

				err := command.Execute(upConfig, incomingState)
				Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
			})

			It("returns an error when the ops file cannot be read", func() {
				err := command.Execute(commands.UpConfig{OpsFile: "some/fake/path"}, storage.State{})
				Expect(err).To(MatchError("Reading ops-file contents: open some/fake/path: no such file or directory"))
			})

			It("returns an error when the env id manager fails", func() {
				envIDManager.SyncCall.Returns.Error = errors.New("apple")

				err := command.Execute(upConfig, incomingState)
				Expect(err).To(MatchError("Env id manager sync: apple"))
			})

			It("returns an error when saving the state fails after env id sync", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("kiwi")}}

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Save state after sync: kiwi"))
			})

			It("returns the error when the terraform manager fails with non terraformManagerError", func() {
				terraformManager.ApplyCall.Returns.Error = errors.New("passionfruit")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				// Expect(err).To(MatchError("Terraform Manager Apply: passionfruit"))
				Expect(err).To(MatchError("passionfruit"))
			})

			It("returns an error when saving the state fails after terraform apply", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {Error: errors.New("kiwi")}}

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Save state after terraform apply: kiwi"))
			})

			It("returns an error when the terraform manager cannot get terraform outputs", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("raspberry")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Parse terraform outputs: raspberry"))
			})

			It("returns an error when the jumpbox cannot be deployed", func() {
				boshManager.CreateJumpboxCall.Returns.Error = errors.New("pineapple")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Create jumpbox: pineapple"))
			})

			It("returns an error when saving the state fails after create jumpbox", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {Error: errors.New("kiwi")}}

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Save state after create jumpbox: kiwi"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshManager.CreateDirectorCall.Returns.Error = errors.New("pineapple")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Create bosh director: pineapple"))
			})

			It("returns an error when saving the state fails after create director", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {Error: errors.New("kiwi")}}

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Save state after create director: kiwi"))
			})

			It("returns an error when the cloud config cannot be uploaded", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("coconut")

				err := command.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Update cloud config: coconut"))
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
					err := command.Execute(commands.UpConfig{}, storage.State{})
					Expect(err).To(MatchError("grapefruit"))

					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives[1].State).To(Equal(partialState))
				})

				It("returns an error when the applier fails and we cannot retrieve the updated bbl state", func() {
					managerError.BBLStateCall.Returns.Error = errors.New("failed to retrieve bbl state")

					err := command.Execute(commands.UpConfig{}, storage.State{})
					Expect(err).To(MatchError("the following errors occurred:\ngrapefruit,\nfailed to retrieve bbl state"))
				})

				Context("when we fail to set the bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.BBLState = partialState
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set bbl state")}}
					})

					It("saves the bbl state and returns the error", func() {
						err := command.Execute(commands.UpConfig{}, storage.State{})
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
					err := command.Execute(commands.UpConfig{}, incomingState)
					Expect(err).To(MatchError("Create bosh director: rambutan"))

					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives[3].State).To(Equal(partialState))
				})

				It("returns a compound error when it fails to save the state", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("lychee")}}

					err := command.Execute(commands.UpConfig{}, incomingState)
					Expect(err).To(MatchError("Save state after bosh director create error: rambutan, lychee"))

					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives[3].State).To(Equal(partialState))
				})
			})
		})
	})
})
