package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-lbs", func() {
	var (
		command              commands.CreateLBs
		boshManager          *fakes.BOSHManager
		lbArgsHandler        *fakes.LBArgsHandler
		logger               *fakes.Logger
		stateValidator       *fakes.StateValidator
		terraformManager     *fakes.TerraformManager
		stateStore           *fakes.StateStore
		environmentValidator *fakes.EnvironmentValidator
		cloudConfigManager   *fakes.CloudConfigManager

		lbState       storage.LB
		mergedLBState storage.LB
	)

	BeforeEach(func() {
		boshManager = &fakes.BOSHManager{}
		lbArgsHandler = &fakes.LBArgsHandler{}
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		stateStore = &fakes.StateStore{}
		environmentValidator = &fakes.EnvironmentValidator{}

		boshManager.VersionCall.Returns.Version = "2.0.24"
		lbState = storage.LB{
			Domain: "something.io",
		}
		mergedLBState = storage.LB{
			Type:   "some type",
			Domain: "something.io",
		}
		lbArgsHandler.GetLBStateCall.Returns.LB = lbState
		lbArgsHandler.MergeCall.Returns.LB = mergedLBState

		command = commands.NewCreateLBs(
			logger,
			stateValidator,
			boshManager,
			lbArgsHandler,
			cloudConfigManager,
			terraformManager,
			stateStore,
			environmentValidator,
		)
	})

	Describe("CheckFastFails", func() {
		Context("when state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("raspberry")
			})

			It("returns an error", func() {
				err := command.CheckFastFails([]string{"--type", "concourse"}, storage.State{IAAS: "gcp"})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("Validate state: raspberry"))
			})
		})

		Context("when the BOSH version is less than 2.0.24 and there is a director", func() {
			It("returns a helpful error message", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{
					"--type", "concourse",
				}, storage.State{
					IAAS:       "gcp",
					NoDirector: false,
				})
				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})
		})

		Context("when the BOSH version is less than 2.0.24 and there is no director", func() {
			It("does not fast fail", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{"--type", "concourse"}, storage.State{
					IAAS:       "gcp",
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when lb args validator fails", func() {
			It("returns an error", func() {
				lbArgsHandler.GetLBStateCall.Returns.Error = errors.New("failed to validate")
				err := command.CheckFastFails([]string{
					"--type", "concourse",
					"--cert", "/path/to/cert",
					"--key", "/path/to/key",
					"--chain", "/path/to/chain",
				}, storage.State{
					IAAS: "aws",
				})

				Expect(err).To(MatchError("failed to validate"))
				Expect(lbArgsHandler.GetLBStateCall.Receives.IAAS).To(Equal("aws"))
				Expect(lbArgsHandler.GetLBStateCall.Receives.Config.LBType).To(Equal("concourse"))
				Expect(lbArgsHandler.GetLBStateCall.Receives.Config.CertPath).To(Equal("/path/to/cert"))
				Expect(lbArgsHandler.GetLBStateCall.Receives.Config.KeyPath).To(Equal("/path/to/key"))
				Expect(lbArgsHandler.GetLBStateCall.Receives.Config.ChainPath).To(Equal("/path/to/chain"))
			})
		})
	})

	Describe("Execute", func() {
		var (
			oldLB               storage.LB
			incomingState       storage.State
			mergedState         storage.State
			terraformApplyState storage.State
		)

		BeforeEach(func() {
			oldLB = storage.LB{Type: "existing-type"}
			incomingState = storage.State{
				IAAS: "aws",
				LB:   oldLB,
			}
			mergedState = incomingState
			mergedState.LB = mergedLBState
			terraformApplyState = storage.State{LatestTFOutput: "terraform-apply"}

			terraformManager.ApplyCall.Returns.BBLState = terraformApplyState
		})

		It("handles arguments, calls terraform, and updates cloud config", func() {
			err := command.Execute([]string{
				"--type", "new-type",
				"--cert", "my-cert",
				"--key", "my-key",
				"--chain", "my-chain",
				"--domain", "my-domain",
			}, incomingState)

			Expect(err).NotTo(HaveOccurred())

			Expect(lbArgsHandler.GetLBStateCall.Receives.IAAS).To(Equal("aws"))
			Expect(lbArgsHandler.GetLBStateCall.Receives.Config).To(Equal(commands.CreateLBsConfig{
				LBType:    "new-type",
				CertPath:  "my-cert",
				KeyPath:   "my-key",
				ChainPath: "my-chain",
				Domain:    "my-domain",
			}))

			Expect(terraformManager.ValidateVersionCall.CallCount).To(Equal(1))

			Expect(lbArgsHandler.MergeCall.Receives.Old).To(Equal(oldLB))
			Expect(lbArgsHandler.MergeCall.Receives.New).To(Equal(lbState))

			Expect(environmentValidator.ValidateCall.Receives.State).To(Equal(mergedState))

			Expect(stateStore.SetCall.Receives[0].State).To(Equal(mergedState))

			Expect(terraformManager.InitCall.CallCount).To(Equal(1))
			Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(mergedState))

			Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(mergedState))

			Expect(stateStore.SetCall.Receives[1].State).To(Equal(terraformApplyState))

			Expect(cloudConfigManager.InitializeCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.InitializeCall.Receives.State).To(Equal(terraformApplyState))
			Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.UpdateCall.Receives.State).To(Equal(terraformApplyState))
		})

		Context("on gcp", func() {
			It("does not accept a chain argument", func() {
				err := command.Execute([]string{"--chain", "some-chain"}, storage.State{IAAS: "gcp"})
				Expect(err).To(MatchError("flag provided but not defined: -chain"))
			})
		})

		Context("no director", func() {
			BeforeEach(func() {
				terraformApplyState.NoDirector = true
				terraformManager.ApplyCall.Returns.BBLState = terraformApplyState
			})

			It("handles arguments and calls terraform, but does not update the cloud config", func() {
				err := command.Execute([]string{
					"--type", "new-type",
					"--cert", "my-cert",
					"--key", "my-key",
					"--chain", "my-chain",
					"--domain", "my-domain",
				}, incomingState)

				Expect(err).NotTo(HaveOccurred())

				Expect(lbArgsHandler.GetLBStateCall.Receives.IAAS).To(Equal("aws"))
				Expect(lbArgsHandler.GetLBStateCall.Receives.Config).To(Equal(commands.CreateLBsConfig{
					LBType:    "new-type",
					CertPath:  "my-cert",
					KeyPath:   "my-key",
					ChainPath: "my-chain",
					Domain:    "my-domain",
				}))

				Expect(terraformManager.ValidateVersionCall.CallCount).To(Equal(1))

				Expect(lbArgsHandler.MergeCall.Receives.Old).To(Equal(oldLB))
				Expect(lbArgsHandler.MergeCall.Receives.New).To(Equal(lbState))

				Expect(environmentValidator.ValidateCall.Receives.State).To(Equal(mergedState))

				Expect(stateStore.SetCall.Receives[0].State).To(Equal(mergedState))

				Expect(terraformManager.InitCall.CallCount).To(Equal(1))
				Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(mergedState))

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformManager.ApplyCall.Receives.BBLState).To(Equal(mergedState))

				Expect(stateStore.SetCall.Receives[1].State).To(Equal(terraformApplyState))

				Expect(cloudConfigManager.InitializeCall.CallCount).To(Equal(0))
				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when an invalid command line flag is supplied", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--invalid-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
				})
			})

			Context("when terraform version validation fails", func() {
				BeforeEach(func() {
					terraformManager.ValidateVersionCall.Returns.Error = errors.New("an error")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("an error"))
				})
			})

			Context("when lb args validation fails", func() {
				BeforeEach(func() {
					lbArgsHandler.GetLBStateCall.Returns.Error = errors.New("some error")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("some error"))
				})
			})

			Context("when environment validation fails", func() {
				BeforeEach(func() {
					environmentValidator.ValidateCall.Returns.Error = errors.New("an error")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("an error"))
				})
			})

			Context("when initial state storage fails", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{fakes.SetCallReturn{Error: errors.New("an error")}}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("saving state before terraform init: an error"))
				})
			})

			Context("when terraform initialization fails", func() {
				BeforeEach(func() {
					terraformManager.InitCall.Returns.Error = errors.New("an error")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("an error"))
				})
			})

			Context("when terraform apply fails", func() {
				BeforeEach(func() {
					terraformManager.ApplyCall.Returns.Error = errors.New("an error")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("an error"))
				})
			})

			Context("when second state storage fails", func() {
				BeforeEach(func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{
						fakes.SetCallReturn{},
						fakes.SetCallReturn{Error: errors.New("an error")},
					}
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("saving state after terraform apply: an error"))
				})
			})

			Context("when cloud config initialize fails", func() {
				BeforeEach(func() {
					cloudConfigManager.InitializeCall.Returns.Error = errors.New("an error")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("an error"))
				})
			})

			Context("when cloud config update fails", func() {
				BeforeEach(func() {
					cloudConfigManager.UpdateCall.Returns.Error = errors.New("an error")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("an error"))
				})
			})
		})
	})
})
