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
		createLBsCmd         *fakes.CreateLBsCmd
		boshManager          *fakes.BOSHManager
		certificateValidator *fakes.CertificateValidator
		logger               *fakes.Logger
		stateValidator       *fakes.StateValidator
	)

	BeforeEach(func() {
		createLBsCmd = &fakes.CreateLBsCmd{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"
		certificateValidator = &fakes.CertificateValidator{}
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}

		command = commands.NewCreateLBs(createLBsCmd, logger, stateValidator, certificateValidator, boshManager)
	})

	Describe("CheckFastFails", func() {
		Context("when state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("raspberry")
			})

			It("returns an error", func() {
				err := command.CheckFastFails([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("Validate state: raspberry"))
			})
		})

		Context("if there is no lb type", func() {
			It("returns an error", func() {
				err := command.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("--type is required"))
			})
		})

		Context("if there is an lb type in the state file", func() {
			It("does not return an error", func() {
				err := command.CheckFastFails([]string{}, storage.State{
					IAAS: "gcp",
					LB: storage.LB{
						Type: "concourse",
					},
				})
				Expect(err).NotTo(HaveOccurred())
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
					"--type", "concourse",
					"--cert", "/path/to/cert",
					"--key", "/path/to/key",
					"--chain", "/path/to/chain",
				}, storage.State{
					IAAS: "aws",
				})

				Expect(err).To(MatchError("Validate certificate: failed to validate"))
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

		Context("when lb type is concourse and domain flag is supplied", func() {
			It("returns an error", func() {
				err := command.CheckFastFails(
					[]string{
						"--type", "concourse",
						"--domain", "ci.example.com",
					},
					storage.State{
						IAAS: "gcp",
					})
				Expect(err).To(MatchError("--domain is not implemented for concourse load balancers. Remove the --domain flag and try again."))
			})
		})
	})

	Describe("Execute", func() {
		Context("if the iaas if GCP", func() {
			It("creates a GCP lb type", func() {
				err := command.Execute([]string{
					"--type", "concourse",
				}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(createLBsCmd.ExecuteCall.Receives.Config).Should(Equal(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType: "concourse",
				}}))
			})
		})

		Context("if GCP and type is cf", func() {
			It("creates a GCP cf lb type is the iaas", func() {
				err := command.Execute([]string{
					"--type", "cf",
					"--cert", "my-cert",
					"--key", "my-key",
					"--domain", "some-domain",
				}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(createLBsCmd.ExecuteCall.Receives.Config).Should(Equal(commands.CreateLBsConfig{GCP: commands.GCPCreateLBsConfig{
					LBType:   "cf",
					CertPath: "my-cert",
					KeyPath:  "my-key",
					Domain:   "some-domain",
				}}))
			})
		})

		Context("if the iaas is AWS", func() {
			It("creates an AWS lb type", func() {
				err := command.Execute([]string{
					"--type", "concourse",
					"--cert", "my-cert",
					"--key", "my-key",
					"--chain", "my-chain",
					"--domain", "some-domain",
				}, storage.State{
					IAAS: "aws",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(createLBsCmd.ExecuteCall.Receives.Config).Should(Equal(
					commands.CreateLBsConfig{
						AWS: commands.AWSCreateLBsConfig{
							LBType:    "concourse",
							CertPath:  "my-cert",
							KeyPath:   "my-key",
							ChainPath: "my-chain",
							Domain:    "some-domain",
						},
					},
				))
			})
		})

		Context("when an LB already exists", func() {
			Context("using GCP", func() {
				It("creates a GCP lb using the existing LB type", func() {
					err := command.Execute([]string{
						"--cert", "some-new-cert",
						"--key", "some-new-key",
					}, storage.State{
						IAAS: "gcp",
						LB: storage.LB{
							Type: "cf",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(createLBsCmd.ExecuteCall.Receives.Config).Should(Equal(
						commands.CreateLBsConfig{
							GCP: commands.GCPCreateLBsConfig{
								LBType:   "cf",
								CertPath: "some-new-cert",
								KeyPath:  "some-new-key",
							},
						},
					))
				})
			})

			Context("using AWS", func() {
				It("creates an AWS lb using the existing LB type", func() {
					err := command.Execute([]string{
						"--cert", "some-new-cert",
						"--key", "some-new-key",
					}, storage.State{
						IAAS: "aws",
						LB: storage.LB{
							Type: "concourse",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(createLBsCmd.ExecuteCall.Receives.Config).Should(Equal(
						commands.CreateLBsConfig{
							AWS: commands.AWSCreateLBsConfig{
								LBType:   "concourse",
								CertPath: "some-new-cert",
								KeyPath:  "some-new-key",
							},
						},
					))
				})
			})
		})

		Context("failure cases", func() {
			Context("when an invalid command line flag is supplied", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--invalid-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
				})
			})

			Context("when the AWSCreateLBs fails", func() {
				BeforeEach(func() {
					createLBsCmd.ExecuteCall.Returns.Error = errors.New("something bad happened")
				})

				It("returns an error", func() {
					err := command.Execute([]string{"some-aws-args"}, storage.State{
						IAAS: "aws",
					})
					Expect(err).To(MatchError("something bad happened"))
				})
			})

			Context("when the GCPCreateLBs fails", func() {
				BeforeEach(func() {
					createLBsCmd.ExecuteCall.Returns.Error = errors.New("something bad happened")
				})

				It("returns an error", func() {
					err := command.Execute([]string{"some-gcp-args"}, storage.State{
						IAAS: "gcp",
					})
					Expect(err).To(MatchError("something bad happened"))
				})
			})
		})
	})
})
