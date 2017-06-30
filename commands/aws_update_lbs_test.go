package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWS Update LBs", func() {
	var (
		credentialValidator  *fakes.CredentialValidator
		environmentValidator *fakes.EnvironmentValidator
		awsCreateLBs         *fakes.AWSCreateLBs

		command commands.AWSUpdateLBs

		incomingState storage.State
	)

	BeforeEach(func() {
		credentialValidator = &fakes.CredentialValidator{}
		environmentValidator = &fakes.EnvironmentValidator{}
		awsCreateLBs = &fakes.AWSCreateLBs{}

		incomingState = storage.State{
			IAAS:    "aws",
			TFState: "some-tf-state",
			LB: storage.LB{
				Type:   "cf",
				Cert:   "some-cert",
				Key:    "some-key",
				Domain: "some-domain",
			},
		}

		command = commands.NewAWSUpdateLBs(awsCreateLBs, credentialValidator, environmentValidator)
	})

	Describe("Execute", func() {
		It("calls out to AWS Create LBs", func() {
			config := commands.AWSCreateLBsConfig{
				CertPath: "some-cert-path",
				KeyPath:  "some-key-path",
				LBType:   "cf",
				Domain:   "some-domain",
			}
			err := command.Execute(config, incomingState)

			Expect(err).NotTo(HaveOccurred())
			Expect(awsCreateLBs.ExecuteCall.CallCount).To(Equal(1))
			Expect(awsCreateLBs.ExecuteCall.Receives.Config).To(Equal(commands.AWSCreateLBsConfig{
				CertPath: "some-cert-path",
				KeyPath:  "some-key-path",
				LBType:   "cf",
				Domain:   "some-domain",
			}))
			Expect(awsCreateLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
		})

		Context("when config does not contain system domain", func() {
			It("passes system domain from the state", func() {
				config := commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "cf",
					Domain:   "",
				}
				err := command.Execute(config, incomingState)

				Expect(err).NotTo(HaveOccurred())
				Expect(awsCreateLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(awsCreateLBs.ExecuteCall.Receives.Config).To(Equal(commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "cf",
					Domain:   "some-domain",
				}))
				Expect(awsCreateLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when config does not contain lb type", func() {
			It("passes lb type from the state", func() {
				config := commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "",
					Domain:   "some-domain",
				}
				err := command.Execute(config, incomingState)

				Expect(err).NotTo(HaveOccurred())
				Expect(awsCreateLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(awsCreateLBs.ExecuteCall.Receives.Config).To(Equal(commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "cf",
					Domain:   "some-domain",
				}))
				Expect(awsCreateLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when an error occurs", func() {
			Context("when credential validation fails", func() {
				It("returns an error", func() {
					credentialValidator.ValidateCall.Returns.Error = errors.New("aws credentials validator failed")

					err := command.Execute(commands.AWSCreateLBsConfig{}, storage.State{})

					Expect(err).To(MatchError("aws credentials validator failed"))
				})
			})

			Context("when environment validation fails", func() {
				It("returns an error", func() {
					environmentValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
					err := command.Execute(commands.AWSCreateLBsConfig{}, incomingState)

					Expect(err).To(MatchError("failed to validate"))
					Expect(environmentValidator.ValidateCall.Receives.State).To(Equal(incomingState))
					Expect(environmentValidator.ValidateCall.CallCount).To(Equal(1))
				})
			})
		})
	})
})
