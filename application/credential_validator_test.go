package application_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialValidator", func() {
	var credentialValidator application.CredentialValidator

	Describe("ValidateCredentials", func() {
		It("validates that the aws credentials have been set", func() {
			credentialValidator = application.NewCredentialValidator(application.Configuration{
				State: storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
				},
			})
			err := credentialValidator.ValidateAWS()
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates that the gcp credentials have been set", func() {
			credentialValidator = application.NewCredentialValidator(application.Configuration{
				State: storage.State{
					GCP: storage.GCP{
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-service-account-key",
						Region:            "some-region",
						Zone:              "some-zone",
					},
				},
			})
			err := credentialValidator.ValidateGCP()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			Context("aws validator", func() {
				It("returns an error when the access key id is missing", func() {
					credentialValidator = application.NewCredentialValidator(application.Configuration{
						State: storage.State{
							AWS: storage.AWS{
								SecretAccessKey: "some-secret-access-key",
								Region:          "some-region",
							},
						},
					})
					Expect(credentialValidator.ValidateAWS()).To(MatchError("AWS access key ID must be provided"))
				})

				It("returns an error when the secret access key is missing", func() {
					credentialValidator = application.NewCredentialValidator(application.Configuration{
						State: storage.State{
							AWS: storage.AWS{
								AccessKeyID: "some-access-key-id",
								Region:      "some-region",
							},
						},
					})
					Expect(credentialValidator.ValidateAWS()).To(MatchError("AWS secret access key must be provided"))
				})

				It("returns an error when the region is missing", func() {
					credentialValidator = application.NewCredentialValidator(application.Configuration{
						State: storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "some-access-key-id",
								SecretAccessKey: "some-secret-access-key",
							},
						},
					})
					Expect(credentialValidator.ValidateAWS()).To(MatchError("AWS region must be provided"))
				})
			})

			Context("gcp validator", func() {
				It("returns an error when the project id is missing", func() {
					credentialValidator = application.NewCredentialValidator(application.Configuration{
						State: storage.State{
							GCP: storage.GCP{
								ServiceAccountKey: "some-service-account-key",
								Region:            "some-region",
								Zone:              "some-zone",
							},
						},
					})
					Expect(credentialValidator.ValidateGCP()).To(MatchError("GCP project ID must be provided"))
				})

				It("returns an error when the service account key is missing", func() {
					credentialValidator = application.NewCredentialValidator(application.Configuration{
						State: storage.State{
							GCP: storage.GCP{
								ProjectID: "some-project-id",
								Region:    "some-region",
								Zone:      "some-zone",
							},
						},
					})
					Expect(credentialValidator.ValidateGCP()).To(MatchError("GCP service account key must be provided"))
				})

				It("returns an error when the region is missing", func() {
					credentialValidator = application.NewCredentialValidator(application.Configuration{
						State: storage.State{
							GCP: storage.GCP{
								ProjectID:         "some-project-id",
								ServiceAccountKey: "some-service-account-key",
								Zone:              "some-zone",
							},
						},
					})
					Expect(credentialValidator.ValidateGCP()).To(MatchError("GCP region must be provided"))
				})

				It("returns an error when the zone is missing", func() {
					credentialValidator = application.NewCredentialValidator(application.Configuration{
						State: storage.State{
							GCP: storage.GCP{
								ProjectID:         "some-project-id",
								ServiceAccountKey: "some-service-account-key",
								Region:            "some-region",
							},
						},
					})
					Expect(credentialValidator.ValidateGCP()).To(MatchError("GCP zone must be provided"))
				})

			})
		})
	})
})
