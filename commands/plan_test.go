package commands_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plan", func() {
	var (
		command commands.Plan

		boshManager        *fakes.BOSHManager
		cloudConfigManager *fakes.CloudConfigManager
		envIDManager       *fakes.EnvIDManager
		lbArgsHandler      *fakes.LBArgsHandler
		logger             *fakes.Logger
		stateStore         *fakes.StateStore
		terraformManager   *fakes.TerraformManager

		tempDir string
	)

	BeforeEach(func() {
		boshManager = &fakes.BOSHManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		envIDManager = &fakes.EnvIDManager{}
		lbArgsHandler = &fakes.LBArgsHandler{}
		logger = &fakes.Logger{}
		stateStore = &fakes.StateStore{}
		terraformManager = &fakes.TerraformManager{}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		boshManager.VersionCall.Returns.Version = "2.0.24"
		stateStore.GetBblDirCall.Returns.Directory = tempDir

		command = commands.NewPlan(
			boshManager,
			cloudConfigManager,
			stateStore,
			envIDManager,
			terraformManager,
			lbArgsHandler,
			logger,
		)
	})

	Describe("Execute", func() {
		var (
			state       storage.State
			syncedState storage.State
		)

		BeforeEach(func() {
			state = storage.State{ID: "some-state-id", IAAS: "some-iaas"}
			syncedState = storage.State{ID: "synced-state-id"}
			envIDManager.SyncCall.Returns.State = syncedState
		})

		It("sets up the bbl state dir", func() {
			args := []string{}
			err := command.Execute(args, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(lbArgsHandler.GetLBStateCall.CallCount).To(Equal(0))

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(envIDManager.SyncCall.Receives.State).To(Equal(state))

			Expect(stateStore.SetCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(syncedState))

			Expect(terraformManager.InitCall.CallCount).To(Equal(1))
			Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(syncedState))

			Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeJumpboxCall.Receives.State).To(Equal(syncedState))

			Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeDirectorCall.Receives.State).To(Equal(syncedState))

			Expect(cloudConfigManager.InitializeCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.InitializeCall.Receives.State).To(Equal(syncedState))
		})

		Context("when --no-director is passed", func() {
			It("sets no director on the state", func() {
				envIDManager.SyncCall.Returns.State = storage.State{NoDirector: true}

				err := command.Execute([]string{"--no-director"}, storage.State{NoDirector: false})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(0))
				Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(0))
			})

			Context("but a director already exists", func() {
				It("returns a helpful error", func() {
					err := command.Execute([]string{"--no-director"}, storage.State{
						BOSH: storage.BOSH{
							DirectorUsername: "admin",
						},
					})
					Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
				})
			})
		})

		Context("when lb flags are passed", func() {
			var lb storage.LB
			BeforeEach(func() {
				lb = storage.LB{
					Type: "some-type",
				}
				lbArgsHandler.GetLBStateCall.Returns.LB = lb
			})

			Context("aws", func() {
				It("sets LB args on the state", func() {
					err := command.Execute(
						[]string{
							"--lb-type", "cf",
							"--lb-cert", "cert",
							"--lb-key", "key",
							"--lb-chain", "chain",
							"--lb-domain", "something.io",
						}, storage.State{IAAS: "aws"})
					Expect(err).NotTo(HaveOccurred())
					Expect(lbArgsHandler.GetLBStateCall.CallCount).To(Equal(1))
					Expect(lbArgsHandler.GetLBStateCall.Receives.IAAS).To(Equal("aws"))
					Expect(lbArgsHandler.GetLBStateCall.Receives.Config).To(Equal(commands.CreateLBsConfig{
						LBType:    "cf",
						CertPath:  "cert",
						KeyPath:   "key",
						ChainPath: "chain",
						Domain:    "something.io",
					}))

					Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
					Expect(envIDManager.SyncCall.Receives.State.LB).To(Equal(lb))
				})
			})
		})

		Context("when --ops-file is passed", func() {
			var (
				opsFilePath     string
				opsFileContents string
			)

			BeforeEach(func() {
				opsFile, err := ioutil.TempFile("", "ops-file")
				Expect(err).NotTo(HaveOccurred())

				opsFilePath = opsFile.Name()

				opsFileContents = "some-ops-file-contents"
				err = ioutil.WriteFile(opsFilePath, []byte(opsFileContents), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})
			It("passes the ops file contents to the bosh manager", func() {
				err := command.Execute([]string{"--ops-file", opsFilePath}, storage.State{})

				Expect(err).NotTo(HaveOccurred())
				Expect(boshManager.InitializeDirectorCall.Receives.State.BOSH.UserOpsFile).To(Equal(opsFileContents))
			})
		})

		Describe("failure cases", func() {
			It("returns an error if reading the ops file fails", func() {
				err := command.Execute([]string{"--ops-file", "some-invalid-path"}, storage.State{})
				Expect(err).To(MatchError("Reading ops-file contents: open some-invalid-path: no such file or directory"))
			})

			It("returns an error if state store set fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("peach")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state: peach"))
			})

			It("returns an error if terraform manager init fails", func() {
				terraformManager.InitCall.Returns.Error = errors.New("pomegranate")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Terraform manager init: pomegranate"))
			})

			It("returns an error if bosh manager initialize jumpbox fails", func() {
				boshManager.InitializeJumpboxCall.Returns.Error = errors.New("tomato")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Bosh manager initialize jumpbox: tomato"))
			})

			It("returns an error if bosh manager initialize director fails", func() {
				boshManager.InitializeDirectorCall.Returns.Error = errors.New("tomatoe")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Bosh manager initialize director: tomatoe"))
			})

			It("returns an error if cloud config initialize fails", func() {
				cloudConfigManager.InitializeCall.Returns.Error = errors.New("potato")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Cloud config manager initialize: potato"))
			})
		})
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

	Describe("ParseArgs", func() {
		Context("when the --ops-file flag is specified", func() {
			var providedOpsFilePath string
			BeforeEach(func() {
				opsFileDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				providedOpsFilePath = filepath.Join(opsFileDir, "some-ops-file")

				err = ioutil.WriteFile(providedOpsFilePath, []byte("some-ops-file-contents"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a config with the ops-file path contents", func() {
				config, err := command.ParseArgs([]string{
					"--ops-file", providedOpsFilePath,
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(config.OpsFile).To(Equal("some-ops-file-contents"))

				By("notifying the user the flag is deprecated", func() {
					Expect(logger.PrintlnCall.Receives.Message).To(Equal(`Deprecation warning: the --ops-file flag is now deprecated and will be removed in bbl v6.0.0. Use "bbl plan" and modify create-director.sh in your state directory to supply operations files for bosh-deployment.`))
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

				By("notifying the user the flag is deprecated", func() {
					Expect(logger.PrintlnCall.Receives.Message).To(Equal("Deprecation warning: the --no-director flag has been deprecated."))
				})
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

		Context("when --lb-type is passed", func() {
			var lb storage.LB
			BeforeEach(func() {
				lb = storage.LB{
					Type: "some-type",
				}
				lbArgsHandler.GetLBStateCall.Returns.LB = lb
			})

			Context("aws", func() {
				It("sets LB args on the state", func() {
					config, err := command.ParseArgs(
						[]string{
							"--lb-type", "cf",
							"--lb-cert", "cert",
							"--lb-key", "key",
							"--lb-chain", "chain",
							"--lb-domain", "something.io",
						}, storage.State{IAAS: "aws"})
					Expect(err).NotTo(HaveOccurred())
					Expect(lbArgsHandler.GetLBStateCall.CallCount).To(Equal(1))
					Expect(lbArgsHandler.GetLBStateCall.Receives.IAAS).To(Equal("aws"))
					Expect(lbArgsHandler.GetLBStateCall.Receives.Config).To(Equal(commands.CreateLBsConfig{
						LBType:    "cf",
						CertPath:  "cert",
						KeyPath:   "key",
						ChainPath: "chain",
						Domain:    "something.io",
					}))

					Expect(config.LB).To(Equal(lb))
				})
			})

			Context("gcp", func() {
				It("doesn't use --lb-chain", func() {
					_, err := command.ParseArgs(
						[]string{
							"--lb-chain", "chain",
						}, storage.State{IAAS: "gcp"})
					Expect(err).To(MatchError("flag provided but not defined: -lb-chain"))
				})
			})

			Context("when the lb args are not valid", func() {
				BeforeEach(func() {
					lbArgsHandler.GetLBStateCall.Returns.Error = errors.New("banana")
				})
				It("returns an error", func() {
					_, err := command.ParseArgs(
						[]string{
							"--lb-type", "cf",
							"--lb-cert", "cert",
							"--lb-key", "key",
							"--lb-domain", "something.io",
						}, storage.State{IAAS: "gcp"})
					Expect(err).To(MatchError("banana"))
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

	Describe("IsInitialized", func() {
		var incomingState storage.State
		BeforeEach(func() {
			incomingState = storage.State{}
		})

		Context("when the state is empty", func() {
			It("returns false", func() {
				Expect(command.IsInitialized(incomingState)).To(BeFalse())
			})
		})

		Context("when the state is old", func() {
			BeforeEach(func() {
				incomingState.BBLVersion = "5.1.3"
			})

			It("returns false", func() {
				Expect(command.IsInitialized(incomingState)).To(BeFalse())
			})
		})

		Context("when the state is new", func() {
			BeforeEach(func() {
				incomingState.BBLVersion = "5.2.0"
			})

			It("returns true", func() {
				Expect(command.IsInitialized(incomingState)).To(BeTrue())
			})
		})

		Context("when the state is from a dev version", func() {
			BeforeEach(func() {
				incomingState.BBLVersion = "5.2.0"
			})

			It("returns true", func() {
				Expect(command.IsInitialized(incomingState)).To(BeTrue())
			})
		})
	})
})
