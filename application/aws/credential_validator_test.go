package aws_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/application/aws"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialValidator", func() {
	var credentialValidator aws.CredentialValidator

	Describe("Validate", func() {
		It("validates that the aws credentials have been set", func() {
			credentialValidator = aws.NewCredentialValidator(application.Configuration{
				State: storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
				},
			})
			err := credentialValidator.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the access key id is missing", func() {
				credentialValidator = aws.NewCredentialValidator(application.Configuration{
					State: storage.State{
						AWS: storage.AWS{
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
						},
					},
				})
				Expect(credentialValidator.Validate()).To(MatchError("AWS access key ID must be provided"))
			})

			It("returns an error when the secret access key is missing", func() {
				credentialValidator = aws.NewCredentialValidator(application.Configuration{
					State: storage.State{
						AWS: storage.AWS{
							AccessKeyID: "some-access-key-id",
							Region:      "some-region",
						},
					},
				})
				Expect(credentialValidator.Validate()).To(MatchError("AWS secret access key must be provided"))
			})

			It("returns an error when the region is missing", func() {
				credentialValidator = aws.NewCredentialValidator(application.Configuration{
					State: storage.State{
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
						},
					},
				})
				Expect(credentialValidator.Validate()).To(MatchError("AWS region must be provided"))
			})
		})
	})
})
