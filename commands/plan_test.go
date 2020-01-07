package commands_test

import (
	"errors"
	"os"

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

		boshManager          *fakes.BOSHManager
		cloudConfigManager   *fakes.CloudConfigManager
		runtimeConfigManager *fakes.RuntimeConfigManager
		envIDManager         *fakes.EnvIDManager
		lbArgsHandler        *fakes.LBArgsHandler
		logger               *fakes.Logger
		stateStore           *fakes.StateStore
		terraformManager     *fakes.TerraformManager
		patchDetector        *fakes.PatchDetector
		bblVersion           string
	)

	BeforeEach(func() {
		boshManager = &fakes.BOSHManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		runtimeConfigManager = &fakes.RuntimeConfigManager{}
		envIDManager = &fakes.EnvIDManager{}
		lbArgsHandler = &fakes.LBArgsHandler{}
		logger = &fakes.Logger{}
		stateStore = &fakes.StateStore{}
		terraformManager = &fakes.TerraformManager{}
		patchDetector = &fakes.PatchDetector{}
		bblVersion = "42.0.0"

		boshManager.VersionCall.Returns.Version = "2.0.48"

		command = commands.NewPlan(
			boshManager,
			cloudConfigManager,
			runtimeConfigManager,
			stateStore,
			patchDetector,
			envIDManager,
			terraformManager,
			lbArgsHandler,
			logger,
			bblVersion,
		)
	})

	Describe("Execute", func() {
		var (
			state            storage.State
			stateWithVersion storage.State
			syncedState      storage.State
		)

		BeforeEach(func() {
			state = storage.State{ID: "some-state-id", IAAS: "some-iaas"}
			stateWithVersion = storage.State{ID: "some-state-id", IAAS: "some-iaas", BBLVersion: "42.0.0"}
			syncedState = storage.State{ID: "synced-state-id"}
			envIDManager.SyncCall.Returns.State = syncedState
		})

		It("sets up the bbl state dir", func() {
			args := []string{}
			err := command.Execute(args, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(lbArgsHandler.GetLBStateCall.CallCount).To(Equal(0))

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(envIDManager.SyncCall.Receives.State).To(Equal(stateWithVersion))

			Expect(stateStore.SetCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(syncedState))

			Expect(terraformManager.SetupCall.CallCount).To(Equal(1))
			Expect(terraformManager.SetupCall.Receives.BBLState).To(Equal(syncedState))

			Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeJumpboxCall.Receives.State).To(Equal(syncedState))

			Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeDirectorCall.Receives.State).To(Equal(syncedState))

			Expect(cloudConfigManager.InitializeCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.InitializeCall.Receives.State).To(Equal(syncedState))

			Expect(runtimeConfigManager.InitializeCall.CallCount).To(Equal(1))
			Expect(runtimeConfigManager.InitializeCall.Receives.State).To(Equal(syncedState))

			Expect(patchDetector.FindCall.CallCount).To(Equal(1))
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
							"--lb-domain", "something.io",
						}, storage.State{IAAS: "aws"})
					Expect(err).NotTo(HaveOccurred())
					Expect(lbArgsHandler.GetLBStateCall.CallCount).To(Equal(1))
					Expect(lbArgsHandler.GetLBStateCall.Receives.IAAS).To(Equal("aws"))
					Expect(lbArgsHandler.GetLBStateCall.Receives.Args).To(Equal(commands.LBArgs{
						LBType:   "cf",
						CertPath: "cert",
						KeyPath:  "key",
						Domain:   "something.io",
					}))

					Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
					Expect(envIDManager.SyncCall.Receives.State.LB).To(Equal(lb))
				})
			})
		})

		Describe("failure cases", func() {
			It("returns an error if state store set fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("peach")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state: peach"))
			})

			It("returns an error if terraform manager init fails", func() {
				terraformManager.SetupCall.Returns.Error = errors.New("pomegranate")

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

			It("returns an error if runtime config initialize fails", func() {
				runtimeConfigManager.InitializeCall.Returns.Error = errors.New("bell-pepper")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Runtime config manager initialize: bell-pepper"))
			})

			It("prints the error but continues if patch detector fails", func() {
				patchDetector.FindCall.Returns.Error = errors.New("iceburg lettuce")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("Failed to detect patch files: iceburg lettuce\n")))
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
				boshManager.VersionCall.Returns.Error = bosh.NewBOSHVersionError(errors.New("banana"))
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the version of the bosh-cli is lower than 2.0.48", func() {
			BeforeEach(func() {
				boshManager.VersionCall.Returns.Version = "1.9.1"
				boshManager.PathCall.Returns.Path = "/bin/banana"
			})

			It("returns an error", func() {
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})
				Expect(err).To(MatchError("/bin/banana: bosh-cli version must be at least v2.0.48, but found v1.9.1"))
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
		Context("when the user provides the name flag", func() {
			It("passes the name flag in the up config", func() {
				config, err := command.ParseArgs([]string{
					"--name", "a-better-name",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Name).To(Equal("a-better-name"))
			})
		})

		Context("when the user provides the name flag as an environment variable", func() {
			BeforeEach(func() {
				os.Setenv("BBL_ENV_NAME", "a-better-name")
			})

			AfterEach(func() {
				os.Unsetenv("BBL_ENV_NAME")
			})

			It("passes the name flag in the up config", func() {
				config, err := command.ParseArgs([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Name).To(Equal("a-better-name"))
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
							"--lb-domain", "something.io",
						}, storage.State{IAAS: "aws"})
					Expect(err).NotTo(HaveOccurred())
					Expect(lbArgsHandler.GetLBStateCall.CallCount).To(Equal(1))
					Expect(lbArgsHandler.GetLBStateCall.Receives.IAAS).To(Equal("aws"))
					Expect(lbArgsHandler.GetLBStateCall.Receives.Args).To(Equal(commands.LBArgs{
						LBType:   "cf",
						CertPath: "cert",
						KeyPath:  "key",
						Domain:   "something.io",
					}))

					Expect(config.LB).To(Equal(lb))
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
		Context("when the state schema is < 13", func() {
			BeforeEach(func() {
				incomingState = storage.State{Version: 12}
			})

			It("returns false", func() {
				Expect(command.IsInitialized(incomingState)).To(BeFalse())
			})
		})

		Context("when the state schema is >= 13", func() {
			BeforeEach(func() {
				incomingState = storage.State{Version: 13}
			})

			It("returns true", func() {
				Expect(command.IsInitialized(incomingState)).To(BeTrue())
			})
		})
	})
})
