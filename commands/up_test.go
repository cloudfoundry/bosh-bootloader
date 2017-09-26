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

var _ = Describe("Up", func() {
	var (
		command commands.Up

		iaasUp             *fakes.UpCmd
		boshManager        *fakes.BOSHManager
		terraformManager   *fakes.TerraformManager
		cloudConfigManager *fakes.CloudConfigManager
		stateStore         *fakes.StateStore
		envIDManager       *fakes.EnvIDManager
	)

	BeforeEach(func() {
		iaasUp = &fakes.UpCmd{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"

		terraformManager = &fakes.TerraformManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		stateStore = &fakes.StateStore{}
		envIDManager = &fakes.EnvIDManager{}

		command = commands.NewUp(iaasUp, boshManager, cloudConfigManager, stateStore, envIDManager, terraformManager)
	})

	Describe("CheckFastFails", func() {
		Context("when the version of BOSH is a dev build", func() {
			It("does not fail", func() {
				boshManager.VersionCall.Returns.Error = bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))

				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the version of BOSH is lower than 2.0.24", func() {
			It("returns a helpful error message when bbling up with a director", func() {
				boshManager.VersionCall.Returns.Version = "1.9.1"
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})

			Context("when the no-director flag is specified", func() {
				It("does not return an error", func() {
					boshManager.VersionCall.Returns.Version = "1.9.1"
					err := command.CheckFastFails([]string{
						"--no-director",
					}, storage.State{Version: 999})

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when the version of BOSH cannot be retrieved", func() {
			It("returns an error", func() {
				boshManager.VersionCall.Returns.Error = errors.New("BOOM")
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err.Error()).To(ContainSubstring("BOOM"))
			})
		})

		Context("when the version of BOSH is invalid", func() {
			It("returns an error", func() {
				boshManager.VersionCall.Returns.Version = "X.5.2"
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err.Error()).To(ContainSubstring("invalid syntax"))
			})
		})

		Context("when bbl-state contains an env-id", func() {
			Context("when the passed in name matches the env-id", func() {
				It("returns no error", func() {
					err := command.CheckFastFails([]string{
						"--name", "some-name",
					}, storage.State{EnvID: "some-name"})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the passed in name does not match the env-id", func() {
				It("returns an error", func() {
					err := command.CheckFastFails([]string{
						"--name", "some-other-name",
					}, storage.State{EnvID: "some-name"})
					Expect(err).To(MatchError("The director name cannot be changed for an existing environment. Current name is some-name."))
				})
			})
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
		)
		BeforeEach(func() {
			incomingState = storage.State{TFState: "incoming-state"}
			iaasState = storage.State{TFState: "iaas-state"}
			iaasUp.ExecuteCall.Returns.State = iaasState

			envIDManagerState = storage.State{TFState: "env-id-sync-call"}
			envIDManager.SyncCall.Returns.State = envIDManagerState

			terraformApplyState = storage.State{TFState: "terraform-apply-call"}
			terraformManager.ApplyCall.Returns.BBLState = terraformApplyState

			createJumpboxState = storage.State{TFState: "create-jumpbox-call"}
			boshManager.CreateJumpboxCall.Returns.State = createJumpboxState

			createDirectorState = storage.State{TFState: "create-director-call"}
			boshManager.CreateDirectorCall.Returns.State = createDirectorState
		})

		It("it works", func() {
			err := command.Execute([]string{"--name", "some-name"}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.ValidateVersionCall.CallCount).To(Equal(1))

			Expect(iaasUp.ExecuteCall.CallCount).To(Equal(1))
			Expect(iaasUp.ExecuteCall.Receives.State).To(Equal(incomingState))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(iaasState))

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(envIDManager.SyncCall.Receives.State).To(Equal(iaasState))
			Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-name"))
			Expect(stateStore.SetCall.Receives[1].State).To(Equal(envIDManagerState))

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(envIDManagerState))
			Expect(stateStore.SetCall.Receives[2].State).To(Equal(terraformApplyState))

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(terraformApplyState))

			Expect(boshManager.CreateJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.CreateJumpboxCall.Receives.State).To(Equal(terraformApplyState))
			Expect(stateStore.SetCall.Receives[3].State).To(Equal(createJumpboxState))

			Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(createJumpboxState))
			Expect(stateStore.SetCall.Receives[4].State).To(Equal(createDirectorState))

			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(createDirectorState))

			Expect(stateStore.SetCall.CallCount).To(Equal(5))
		})

		Context("when the config has ops files", func() {
			var opsFilePath string

			BeforeEach(func() {
				opsFile, err := ioutil.TempFile("", "ops-file")
				Expect(err).NotTo(HaveOccurred())

				opsFilePath = opsFile.Name()
				opsFileContents := "some-ops-file-contents"
				err = ioutil.WriteFile(opsFilePath, []byte(opsFileContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("passes the ops file contents to the bosh manager", func() {
				err := command.Execute([]string{"--ops-file", opsFilePath}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateDirectorCall.Receives.State.BOSH.UserOpsFile).To(Equal("some-ops-file-contents"))
			})
		})

		Context("when --no-director flag is passed", func() {
			It("sets NoDirector to true on the state", func() {
				err := command.Execute([]string{"--no-director"}, storage.State{})
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
				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
				Expect(stateStore.SetCall.Receives[2].State.NoDirector).To(BeTrue())
				Expect(stateStore.SetCall.CallCount).To(Equal(3))
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Describe("failure cases", func() {
			It("returns an error if terraform manager version validator fails", func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("grape")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Terraform validate version: grape"))
			})

			It("returns an error as-is when the iaas up command fails", func() {
				iaasUp.ExecuteCall.Returns.Error = errors.New("tomato")
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("tomato"))
			})

			It("returns an error when saving the state fails after env id sync", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("kiwi")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state after IAAS up: kiwi"))
			})

			It("returns an error when parse args fails", func() {
				err := command.Execute([]string{"--foo"}, storage.State{})
				Expect(err).To(MatchError("flag provided but not defined: -foo"))
			})

			It("fast fails when the config has the no-director flag set and the bbl state has a bosh director", func() {
				iaasUp.ExecuteCall.Returns.State = storage.State{BOSH: storage.BOSH{DirectorName: "some-director"}}

				err := command.Execute([]string{"--no-director"}, incomingState)
				Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
			})

			It("returns an error when the ops file cannot be read", func() {
				err := command.Execute([]string{"--ops-file", "some/fake/path"}, storage.State{})
				Expect(err).To(MatchError("Reading ops-file contents: open some/fake/path: no such file or directory"))
			})

			It("returns an error when the env id manager fails", func() {
				envIDManager.SyncCall.Returns.Error = errors.New("apple")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Env id manager sync: apple"))
			})

			It("returns an error when saving the state fails after env id sync", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {Error: errors.New("kiwi")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state after sync: kiwi"))
			})

			It("returns the error when the terraform manager fails with non terraformManagerError", func() {
				terraformManager.ApplyCall.Returns.Error = errors.New("passionfruit")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("passionfruit"))
			})

			It("returns an error when saving the state fails after terraform apply", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {Error: errors.New("kiwi")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state after terraform apply: kiwi"))
			})

			It("returns an error when the terraform manager cannot get terraform outputs", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("raspberry")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Parse terraform outputs: raspberry"))
			})

			It("returns an error when the jumpbox cannot be deployed", func() {
				boshManager.CreateJumpboxCall.Returns.Error = errors.New("pineapple")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Create jumpbox: pineapple"))
			})

			It("returns an error when saving the state fails after create jumpbox", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {Error: errors.New("kiwi")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state after create jumpbox: kiwi"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshManager.CreateDirectorCall.Returns.Error = errors.New("pineapple")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Create bosh director: pineapple"))
			})

			It("returns an error when saving the state fails after create director", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {}, {Error: errors.New("kiwi")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state after create director: kiwi"))
			})

			It("returns an error when the cloud config cannot be uploaded", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("coconut")

				err := command.Execute([]string{}, storage.State{})
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
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("grapefruit"))

					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives[2].State).To(Equal(partialState))
				})

				It("returns an error when the applier fails and we cannot retrieve the updated bbl state", func() {
					managerError.BBLStateCall.Returns.Error = errors.New("failed to retrieve bbl state")

					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("the following errors occurred:\ngrapefruit,\nfailed to retrieve bbl state"))
				})

				Context("when we fail to set the bbl state", func() {
					BeforeEach(func() {
						managerError.BBLStateCall.Returns.BBLState = partialState
						stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("failed to set bbl state")}}
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

					Expect(stateStore.SetCall.CallCount).To(Equal(5))
					Expect(stateStore.SetCall.Receives[4].State).To(Equal(partialState))
				})

				It("returns a compound error when it fails to save the state", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {}, {errors.New("lychee")}}

					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Save state after bosh director create error: rambutan, lychee"))

					Expect(stateStore.SetCall.CallCount).To(Equal(5))
					Expect(stateStore.SetCall.Receives[4].State).To(Equal(partialState))
				})
			})
		})
	})

	Describe("ParseArgs", func() {
		Context("when the --ops-file flag is specified", func() {
			It("returns a config with the ops-file path", func() {
				config, err := command.ParseArgs([]string{
					"--ops-file", "some-ops-file-path",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(config.OpsFile).To(Equal("some-ops-file-path"))
			})

			Context("when the --ops-file flag is not specified", func() {
				It("creates a default ops-file with the contents of state.BOSH.UserOpsFile", func() {
					config, err := command.ParseArgs([]string{}, storage.State{
						BOSH: storage.BOSH{
							UserOpsFile: "some-ops-file-contents",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					filePath := config.OpsFile
					fileContents, err := ioutil.ReadFile(filePath)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(fileContents)).To(Equal("some-ops-file-contents"))
				})
			})
		})

		Context("when the user provides the name flag", func() {
			It("passes the name flag in the up config", func() {
				config, err := command.ParseArgs([]string{
					"--name", "a-better-name",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Name).To(Equal("a-better-name"))
			})
		})

		Context("when the user provides the no-director flag", func() {
			It("passes NoDirector as true in the up config", func() {
				config, err := command.ParseArgs([]string{
					"--no-director",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(config.NoDirector).To(Equal(true))
			})

			Context("when the --no-director flag was omitted on a subsequent bbl-up", func() {
				It("passes no-director as true in the up config", func() {
					config, err := command.ParseArgs([]string{},
						storage.State{
							IAAS:       "gcp",
							NoDirector: true,
						})
					Expect(err).NotTo(HaveOccurred())
					Expect(config.NoDirector).To(Equal(true))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when undefined flags are passed", func() {
				_, err := command.ParseArgs([]string{"--foo", "bar"}, storage.State{})
				Expect(err).To(MatchError("flag provided but not defined: -foo"))
			})
		})
	})
})
