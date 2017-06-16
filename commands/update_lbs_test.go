package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Update LBs", func() {
	var (
		command              commands.UpdateLBs
		certificateValidator *fakes.CertificateValidator
		stateValidator       *fakes.StateValidator
		boshManager          *fakes.BOSHManager
		logger               *fakes.Logger
		awsUpdateLBs         *fakes.AWSUpdateLBs
		gcpUpdateLBs         *fakes.GCPUpdateLBs
	)

	BeforeEach(func() {
		certificateValidator = &fakes.CertificateValidator{}
		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}
		awsUpdateLBs = &fakes.AWSUpdateLBs{}
		gcpUpdateLBs = &fakes.GCPUpdateLBs{}

		command = commands.NewUpdateLBs(awsUpdateLBs, gcpUpdateLBs, certificateValidator, stateValidator, logger, boshManager)
	})

	Describe("CheckFastFails", func() {
		var (
			incomingState storage.State
		)

		BeforeEach(func() {
			boshManager.VersionCall.Returns.Version = "2.0.0"

			incomingState = storage.State{
				IAAS: "aws",
				Stack: storage.Stack{
					LBType: "concourse",
				},
			}
		})

		It("returns an error when state validator fails", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			err := command.CheckFastFails([]string{}, storage.State{})

			Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
			Expect(err).To(MatchError("state validator failed"))
		})

		It("returns an error when invalid flags are provided", func() {
			err := command.CheckFastFails([]string{
				"--invalid-flag",
			}, storage.State{})

			Expect(err).To(MatchError(ContainSubstring("flag provided but not defined")))
		})

		It("returns an error when the certificate validator fails", func() {
			certificateValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
			err := command.CheckFastFails([]string{
				"--cert", "/path/to/cert",
				"--key", "/path/to/key",
				"--chain", "/path/to/chain",
			}, incomingState)

			Expect(err).To(MatchError("failed to validate"))
			Expect(certificateValidator.ValidateCall.Receives.Command).To(Equal("update-lbs"))
			Expect(certificateValidator.ValidateCall.Receives.CertificatePath).To(Equal("/path/to/cert"))
			Expect(certificateValidator.ValidateCall.Receives.KeyPath).To(Equal("/path/to/key"))
			Expect(certificateValidator.ValidateCall.Receives.ChainPath).To(Equal("/path/to/chain"))
		})

		It("returns an error if there is no lb", func() {
			err := command.CheckFastFails([]string{}, storage.State{
				Stack: storage.Stack{
					LBType: "",
				},
			})
			Expect(err).To(MatchError(commands.LBNotFound))
		})

		Context("when --skip-if-missing is provided", func() {
			It("returns no error when lb does not exist", func() {
				err := command.CheckFastFails([]string{
					"--skip-if-missing",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`no lb type exists, skipping...`))
			})
		})

		Context("when the BOSH version is less than 2.0.0 and there is a director", func() {
			It("returns a helpful error message", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{}, incomingState)
				Expect(err).To(MatchError("BOSH version must be at least v2.0.0"))
			})
		})

		Context("when the BOSH version is less than 2.0.0 and there is no director", func() {
			It("returns no error", func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
				err := command.CheckFastFails([]string{}, storage.State{
					NoDirector: true,
					LB: storage.LB{
						Type: "cf",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Execute", func() {
		It("updates a GCP cf lb type is the iaas if GCP and type is cf", func() {
			err := command.Execute([]string{
				"--cert", "my-cert",
				"--key", "my-key",
				"--domain", "some-domain",
			}, storage.State{
				IAAS: "gcp",
				LB: storage.LB{
					Type: "cf",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(gcpUpdateLBs.ExecuteCall.Receives.Config).To(Equal(commands.GCPCreateLBsConfig{
				LBType:   "cf",
				CertPath: "my-cert",
				KeyPath:  "my-key",
				Domain:   "some-domain",
			}))
		})

		It("creates an AWS lb type if the iaas is AWS", func() {
			err := command.Execute([]string{
				"--cert", "my-cert",
				"--key", "my-key",
				"--chain", "my-chain",
			}, storage.State{
				Stack: storage.Stack{
					LBType: "concourse",
				},
				IAAS: "aws",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(awsUpdateLBs.ExecuteCall.Receives.Config).To(Equal(commands.AWSCreateLBsConfig{
				LBType:    "concourse",
				CertPath:  "my-cert",
				KeyPath:   "my-key",
				ChainPath: "my-chain",
			}))
		})

		Context("when --skip-if-missing is provided", func() {
			It("returns no error when lb does not exist", func() {
				err := command.Execute([]string{
					"--skip-if-missing",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`no lb type exists, skipping...`))
			})
		})

		Describe("failure cases", func() {
			It("returns an error when invalid flags are provided", func() {
				err := command.Execute([]string{
					"--invalid-flag",
				}, storage.State{})

				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined")))
			})
		})
	})
})
