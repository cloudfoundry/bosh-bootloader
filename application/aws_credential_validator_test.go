package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
)

var _ = Describe("AWSCredentialValidator", func() {
	var awsCredentialValidator application.AWSCredentialValidator

	BeforeEach(func() {
		awsCredentialValidator = application.NewAWSCredentialValidator()
	})

	Describe("ValidateCredentials", func() {
		It("validates that the credentials have been set", func() {
			err := awsCredentialValidator.Validate("some-access-key-id", " some-secret-access-key", "some-region")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the access key id is missing", func() {
				Expect(awsCredentialValidator.Validate("", " some-secret-access-key", "some-region")).To(MatchError("--aws-access-key-id must be provided"))
			})

			It("returns an error when the secret access key is missing", func() {
				Expect(awsCredentialValidator.Validate("some-access-key-id", "", "some-region")).To(MatchError("--aws-secret-access-key must be provided"))
			})

			It("returns an error when the region is missing", func() {
				Expect(awsCredentialValidator.Validate("some-access-key-id", " some-secret-access-key", "")).To(MatchError("--aws-region must be provided"))
			})
		})
	})
})
