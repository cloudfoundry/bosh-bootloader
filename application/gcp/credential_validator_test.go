package gcp_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/application/gcp"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialValidator", func() {
	var credentialValidator gcp.CredentialValidator

	Describe("Validate", func() {
		It("validates that the gcp credentials have been set", func() {
			credentialValidator = gcp.NewCredentialValidator(application.Configuration{
				State: storage.State{
					GCP: storage.GCP{
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-service-account-key",
						Region:            "some-region",
						Zone:              "some-zone",
					},
				},
			})
			err := credentialValidator.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the project id is missing", func() {
				credentialValidator = gcp.NewCredentialValidator(application.Configuration{
					State: storage.State{
						GCP: storage.GCP{
							ServiceAccountKey: "some-service-account-key",
							Region:            "some-region",
							Zone:              "some-zone",
						},
					},
				})
				Expect(credentialValidator.Validate()).To(MatchError("GCP project ID must be provided"))
			})

			It("returns an error when the service account key is missing", func() {
				credentialValidator = gcp.NewCredentialValidator(application.Configuration{
					State: storage.State{
						GCP: storage.GCP{
							ProjectID: "some-project-id",
							Region:    "some-region",
							Zone:      "some-zone",
						},
					},
				})
				Expect(credentialValidator.Validate()).To(MatchError("GCP service account key must be provided"))
			})

			It("returns an error when the region is missing", func() {
				credentialValidator = gcp.NewCredentialValidator(application.Configuration{
					State: storage.State{
						GCP: storage.GCP{
							ProjectID:         "some-project-id",
							ServiceAccountKey: "some-service-account-key",
							Zone:              "some-zone",
						},
					},
				})
				Expect(credentialValidator.Validate()).To(MatchError("GCP region must be provided"))
			})

			It("returns an error when the zone is missing", func() {
				credentialValidator = gcp.NewCredentialValidator(application.Configuration{
					State: storage.State{
						GCP: storage.GCP{
							ProjectID:         "some-project-id",
							ServiceAccountKey: "some-service-account-key",
							Region:            "some-region",
						},
					},
				})
				Expect(credentialValidator.Validate()).To(MatchError("GCP zone must be provided"))
			})
		})
	})
})
