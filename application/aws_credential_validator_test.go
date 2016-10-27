package application_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSCredentialValidator", func() {
	var awsCredentialValidator application.AWSCredentialValidator

	BeforeEach(func() {
	})

	Describe("ValidateCredentials", func() {
		It("validates that the credentials have been set", func() {
			awsCredentialValidator = application.NewAWSCredentialValidator(application.Configuration{
				State: storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
				},
			})
			err := awsCredentialValidator.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the access key id is missing", func() {
				awsCredentialValidator = application.NewAWSCredentialValidator(application.Configuration{
					State: storage.State{
						AWS: storage.AWS{
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
						},
					},
				})
				Expect(awsCredentialValidator.Validate()).To(MatchError("AWS access key ID must be provided"))
			})

			It("returns an error when the secret access key is missing", func() {
				awsCredentialValidator = application.NewAWSCredentialValidator(application.Configuration{
					State: storage.State{
						AWS: storage.AWS{
							AccessKeyID: "some-access-key-id",
							Region:      "some-region",
						},
					},
				})
				Expect(awsCredentialValidator.Validate()).To(MatchError("AWS secret access key must be provided"))
			})

			It("returns an error when the region is missing", func() {
				awsCredentialValidator = application.NewAWSCredentialValidator(application.Configuration{
					State: storage.State{
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
						},
					},
				})
				Expect(awsCredentialValidator.Validate()).To(MatchError("AWS region must be provided"))
			})
		})
	})
})
