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
		awsCreateLBs         *fakes.AWSCreateLBs
		gcpCreateLBs         *fakes.GCPCreateLBs
		stateValidator       *fakes.StateValidator
		certificateValidator *fakes.CertificateValidator
		boshManager          *fakes.BOSHManager
	)

	BeforeEach(func() {
		awsCreateLBs = &fakes.AWSCreateLBs{}
		gcpCreateLBs = &fakes.GCPCreateLBs{}
		stateValidator = &fakes.StateValidator{}
		certificateValidator = &fakes.CertificateValidator{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"

		command = commands.NewCreateLBs(awsCreateLBs, gcpCreateLBs, stateValidator, certificateValidator, boshManager)
	})

	Describe("CheckFastFails", func() {
		It("returns an error when state validator fails", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			err := command.CheckFastFails([]string{}, storage.State{})

			Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
			Expect(err).To(MatchError("state validator failed"))
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
				err := command.CheckFastFails([]string{
					"--type", "concourse",
				}, storage.State{
					IAAS:       "gcp",
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when certificate validator fails for cert and key", func() {
			It("returns an error", func() {
				certificateValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
				err := command.CheckFastFails([]string{
					"--cert", "/path/to/cert",
					"--key", "/path/to/key",
					"--chain", "/path/to/chain",
				}, storage.State{})

				Expect(err).To(MatchError("failed to validate"))
				Expect(certificateValidator.ValidateCall.Receives.Command).To(Equal("create-lbs"))
				Expect(certificateValidator.ValidateCall.Receives.CertificatePath).To(Equal("/path/to/cert"))
				Expect(certificateValidator.ValidateCall.Receives.KeyPath).To(Equal("/path/to/key"))
				Expect(certificateValidator.ValidateCall.Receives.ChainPath).To(Equal("/path/to/chain"))
			})
		})

		Context("when iaas is gcp and lb type is concourse", func() {
			It("does not call certificateValidator", func() {
				_ = command.CheckFastFails(
					[]string{
						"--type", "concourse",
					},
					storage.State{
						IAAS: "gcp",
					})

				Expect(certificateValidator.ValidateCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("Execute", func() {
		It("creates a GCP lb type if the iaas if GCP", func() {
			err := command.Execute([]string{
				"--type", "concourse",
				"--skip-if-exists",
			}, storage.State{
				IAAS: "gcp",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(gcpCreateLBs.ExecuteCall.Receives.Config).Should(Equal(commands.GCPCreateLBsConfig{
				LBType:       "concourse",
				SkipIfExists: true,
			}))
		})

		It("creates a GCP cf lb type is the iaas if GCP and type is cf", func() {
			err := command.Execute([]string{
				"--type", "cf",
				"--cert", "my-cert",
				"--key", "my-key",
				"--domain", "some-domain",
				"--skip-if-exists",
			}, storage.State{
				IAAS: "gcp",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(gcpCreateLBs.ExecuteCall.Receives.Config).Should(Equal(commands.GCPCreateLBsConfig{
				LBType:       "cf",
				CertPath:     "my-cert",
				KeyPath:      "my-key",
				Domain:       "some-domain",
				SkipIfExists: true,
			}))
		})

		It("creates an AWS lb type if the iaas is AWS", func() {
			err := command.Execute([]string{
				"--type", "concourse",
				"--cert", "my-cert",
				"--key", "my-key",
				"--chain", "my-chain",
				"--domain", "some-domain",
				"--skip-if-exists", "true",
			}, storage.State{
				IAAS: "aws",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(awsCreateLBs.ExecuteCall.Receives.Config).Should(Equal(commands.AWSCreateLBsConfig{
				LBType:       "concourse",
				CertPath:     "my-cert",
				KeyPath:      "my-key",
				ChainPath:    "my-chain",
				Domain:       "some-domain",
				SkipIfExists: true,
			}))
		})

		Context("failure cases", func() {
			It("returns an error when an invalid command line flag is supplied", func() {
				err := command.Execute([]string{"--invalid-flag"}, storage.State{})
				Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
			})

			It("returns an error when the AWSCreateLBs fails", func() {
				awsCreateLBs.ExecuteCall.Returns.Error = errors.New("something bad happened")

				err := command.Execute([]string{"some-aws-args"}, storage.State{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("something bad happened"))
			})

			It("returns an error when the GCPCreateLBs fails", func() {
				gcpCreateLBs.ExecuteCall.Returns.Error = errors.New("something bad happened")

				err := command.Execute([]string{"some-gcp-args"}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
