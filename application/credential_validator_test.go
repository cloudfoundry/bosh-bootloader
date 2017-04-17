package application_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialValidator", func() {
	Describe("Validate", func() {
		var (
			gcpCredentialValidator *fakes.CredentialValidator
			awsCredentialValidator *fakes.CredentialValidator

			credentialValidator application.CredentialValidator
		)

		BeforeEach(func() {
			gcpCredentialValidator = &fakes.CredentialValidator{}
			awsCredentialValidator = &fakes.CredentialValidator{}

			gcpCredentialValidator.ValidateCall.Returns.Error = errors.New("gcp validation failed")
			awsCredentialValidator.ValidateCall.Returns.Error = errors.New("aws validation failed")
		})

		Context("when iaas is gcp", func() {
			BeforeEach(func() {
				configuration := application.Configuration{
					State: storage.State{
						IAAS: "gcp",
					},
				}

				credentialValidator = application.NewCredentialValidator(configuration, gcpCredentialValidator, awsCredentialValidator)
			})

			It("validates using the gcp credential validator", func() {
				err := credentialValidator.Validate()

				Expect(err).To(MatchError("gcp validation failed"))
				Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(0))
			})
		})

		Context("when iaas is aws", func() {
			BeforeEach(func() {
				configuration := application.Configuration{
					State: storage.State{
						IAAS: "aws",
					},
				}

				credentialValidator = application.NewCredentialValidator(configuration, gcpCredentialValidator, awsCredentialValidator)
			})
			It("validates using the aws credential validator", func() {
				err := credentialValidator.Validate()

				Expect(err).To(MatchError("aws validation failed"))
				Expect(gcpCredentialValidator.ValidateCall.CallCount).To(Equal(0))
			})
		})

		Context("when iaas is invalid", func() {
			BeforeEach(func() {
				configuration := application.Configuration{
					State: storage.State{
						IAAS: "invalid",
					},
				}

				credentialValidator = application.NewCredentialValidator(configuration, gcpCredentialValidator, awsCredentialValidator)
			})

			It("returns a helpful error message", func() {
				err := credentialValidator.Validate()

				Expect(err).To(MatchError(`cannot validate credentials: invalid iaas "invalid"`))
				Expect(gcpCredentialValidator.ValidateCall.CallCount).To(Equal(0))
				Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(0))
			})
		})
	})
})
