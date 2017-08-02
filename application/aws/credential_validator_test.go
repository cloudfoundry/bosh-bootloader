package aws_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application/aws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialValidator", func() {
	var credentialValidator aws.CredentialValidator

	Describe("Validate", func() {
		It("validates that the aws credentials have been set", func() {
			credentialValidator = aws.NewCredentialValidator("some-access-key-id", "some-secret-access-key", "some-region")
			err := credentialValidator.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the access key id is missing", func() {
				credentialValidator = aws.NewCredentialValidator("", "some-secret-access-key", "some-region")
				Expect(credentialValidator.Validate()).To(MatchError("AWS access key ID must be provided"))
			})

			It("returns an error when the secret access key is missing", func() {
				credentialValidator = aws.NewCredentialValidator("some-access-key-id", "", "some-region")
				Expect(credentialValidator.Validate()).To(MatchError("AWS secret access key must be provided"))
			})

			It("returns an error when the region is missing", func() {
				credentialValidator = aws.NewCredentialValidator("some-access-key-id", "some-secret-access-key", "")
				Expect(credentialValidator.Validate()).To(MatchError("AWS region must be provided"))
			})
		})
	})
})
