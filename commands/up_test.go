package commands_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

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

		boshManager        *fakes.BOSHManager
		terraformManager   *fakes.TerraformManager
		cloudConfigManager *fakes.CloudConfigManager
		stateStore         *fakes.StateStore
		envIDManager       *fakes.EnvIDManager

		tempDir string
	)

	BeforeEach(func() {
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

		command = commands.NewUp(boshManager, cloudConfigManager, stateStore, envIDManager, terraformManager)
	})

	Describe("CheckFastFails", func() {
		Context("when terraform manager validate version fails", func() {
			It("returns an error", func() {
				terraformManager.ValidateVersionCall.Returns.Error = errors.New("lychee")

				err := command.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("Terraform manager validate version: lychee"))
			})
		})

		Context("when the version of BOSH is a dev build", func() {
			It("does not fail", func() {
				boshManager.VersionCall.Returns.Error = bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the version of the bosh-cli is lower than 2.0.24", func() {
			Context("when there is a bosh director", func() {
				It("returns an error", func() {
					boshManager.VersionCall.Returns.Version = "1.9.1"
					err := command.CheckFastFails([]string{}, storage.State{Version: 999})

					Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
				})
			})

			Context("when there is no director", func() {
				It("does not return an error", func() {
					boshManager.VersionCall.Returns.Version = "1.9.1"
					err := command.CheckFastFails([]string{"--no-director"}, storage.State{Version: 999})

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when bosh -v fails", func() {
			It("returns an error", func() {
				boshManager.VersionCall.Returns.Error = errors.New("BOOM")
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err.Error()).To(ContainSubstring("BOOM"))
			})
		})

		Context("when bosh -v is invalid", func() {
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

			envIDManagerState = storage.State{TFState: "env-id-sync-call"}
			envIDManager.SyncCall.Returns.State = envIDManagerState

			terraformApplyState = storage.State{TFState: "terraform-apply-call"}
			terraformManager.ApplyCall.Returns.BBLState = terraformApplyState

			createJumpboxState = storage.State{TFState: "create-jumpbox-call"}
			boshManager.CreateJumpboxCall.Returns.State = createJumpboxState

			createDirectorState = storage.State{TFState: "create-director-call"}
			boshManager.CreateDirectorCall.Returns.State = createDirectorState

			terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{
				Map: map[string]interface{}{
					"jumpbox_url": "some-jumpbox-url",
				},
			}

			terraformManager.IsInitializedCall.Returns.IsInitialized = true
		})

		It("it works", func() {
			err := command.Execute([]string{"--name", "some-name"}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(envIDManager.SyncCall.Receives.State).To(Equal(incomingState))
			Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-name"))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(envIDManagerState))

			Expect(terraformManager.InitCall.CallCount).To(Equal(0))

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(envIDManagerState))
			Expect(stateStore.SetCall.Receives[1].State).To(Equal(terraformApplyState))

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(terraformApplyState))

			Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeJumpboxCall.Receives.State).To(Equal(terraformApplyState))
			Expect(boshManager.CreateJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.CreateJumpboxCall.Receives.State).To(Equal(terraformApplyState))
			Expect(boshManager.CreateJumpboxCall.Receives.JumpboxURL).To(Equal("some-jumpbox-url"))
			Expect(stateStore.SetCall.Receives[2].State).To(Equal(createJumpboxState))

			Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeDirectorCall.Receives.State).To(Equal(createJumpboxState))
			Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(createJumpboxState))
			Expect(stateStore.SetCall.Receives[3].State).To(Equal(createDirectorState))

			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(createDirectorState))

			Expect(stateStore.SetCall.CallCount).To(Equal(4))
		})

		Context("when terraform is not initialized yet", func() {
			BeforeEach(func() {
				terraformManager.IsInitializedCall.Returns.IsInitialized = false
			})
			It("calls init on the manager", func() {
				err := command.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(terraformManager.InitCall.CallCount).To(Equal(1))
				Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(envIDManagerState))
			})
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

				Expect(boshManager.InitializeDirectorCall.Receives.State.BOSH.UserOpsFile).To(Equal("some-ops-file-contents"))
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
				Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(0))
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[1].State.NoDirector).To(BeTrue())
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Describe("failure cases", func() {
			Context("when parse args fails", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--foo"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -foo"))
				})
			})

			Context("when the config has the no-director flag set and the bbl state has a bosh director", func() {
				BeforeEach(func() {
					incomingState = storage.State{BOSH: storage.BOSH{DirectorName: "some-director"}}
				})

				It("fast fails", func() {
					err := command.Execute([]string{"--no-director"}, incomingState)
					Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
				})
			})

			Context("when the ops file cannot be read", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--ops-file", "some/fake/path"}, storage.State{})
					Expect(err).To(MatchError("Reading ops-file contents: open some/fake/path: no such file or directory"))
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

			Context("when the terraform manager fails on init", func() {
				BeforeEach(func() {
					terraformManager.IsInitializedCall.Returns.IsInitialized = false
					terraformManager.InitCall.Returns.Error = errors.New("grapefruit")
				})

				It("returns the error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Terraform manager init: grapefruit"))
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

			Context("when the jumpbox cannot be initialized", func() {
				BeforeEach(func() {
					boshManager.InitializeJumpboxCall.Returns.Error = errors.New("pineapple")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Create jumpbox: pineapple"))
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

			Context("when bosh cannot be initialized", func() {
				BeforeEach(func() {
					boshManager.InitializeDirectorCall.Returns.Error = errors.New("pineapple")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("Create bosh director: pineapple"))
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
		Context("when the --ops-file flag is specified", func() {
			var providedOpsFilePath string
			BeforeEach(func() {
				opsFileDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				providedOpsFilePath = filepath.Join(opsFileDir, "some-ops-file")

				err = ioutil.WriteFile(providedOpsFilePath, []byte("some-ops-file-contents"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a config with the ops-file path", func() {
				config, err := command.ParseArgs([]string{
					"--ops-file", providedOpsFilePath,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(config.OpsFile).To(Equal(providedOpsFilePath))
			})
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

			It("writes the previous user ops file to the .bbl directory", func() {
				config, err := command.ParseArgs([]string{}, storage.State{
					BOSH: storage.BOSH{
						UserOpsFile: "some-ops-file-contents",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				filePath := config.OpsFile
				fileContents, err := ioutil.ReadFile(filePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(filePath).To(Equal(filepath.Join(tempDir, "previous-user-ops-file.yml")))
				Expect(string(fileContents)).To(Equal("some-ops-file-contents"))
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
			Context("when undefined flags are passed", func() {
				It("returns an error", func() {
					_, err := command.ParseArgs([]string{"--foo", "bar"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -foo"))
				})
			})
		})
	})
})
