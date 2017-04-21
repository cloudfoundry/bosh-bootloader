package application_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvironmentValidator", func() {
	Describe("Validate", func() {
		var (
			gcpEnvironmentValidator *fakes.EnvironmentValidator
			awsEnvironmentValidator *fakes.EnvironmentValidator

			environmentValidator application.EnvironmentValidator
			state                storage.State
		)

		BeforeEach(func() {
			gcpEnvironmentValidator = &fakes.EnvironmentValidator{}
			awsEnvironmentValidator = &fakes.EnvironmentValidator{}

			gcpEnvironmentValidator.ValidateCall.Returns.Error = errors.New("gcp environment validation failed")
			awsEnvironmentValidator.ValidateCall.Returns.Error = errors.New("aws environment validation failed")

			environmentValidator = application.NewEnvironmentValidator(awsEnvironmentValidator, gcpEnvironmentValidator)
		})

		Context("when the IAAS is gcp", func() {
			BeforeEach(func() {
				state = storage.State{
					IAAS: "gcp",
				}
			})

			It("calls the validate function of gcpEnvironmentValidator", func() {
				err := environmentValidator.Validate(state)
				Expect(err).To(MatchError("gcp environment validation failed"))
			})
		})

		Context("when the IAAS is aws", func() {
			BeforeEach(func() {
				state = storage.State{
					IAAS: "aws",
				}
			})

			It("calls the validate function of awsEnvironmentValidator", func() {
				err := environmentValidator.Validate(state)
				Expect(err).To(MatchError("aws environment validation failed"))
			})
		})

		Context("when the IAAS is invalid", func() {
			BeforeEach(func() {
				state = storage.State{
					IAAS: "invalid",
				}
			})

			It("returns an error", func() {
				err := environmentValidator.Validate(state)
				Expect(err).To(MatchError("invalid IAAS specified: invalid"))
			})
		})
	})
})
