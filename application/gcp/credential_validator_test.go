package gcp_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialValidator", func() {
	var credentialValidator gcp.CredentialValidator

	Describe("Validate", func() {
		It("validates that the gcp credentials have been set", func() {
			credentialValidator = gcp.NewCredentialValidator("some-project-id", "some-service-account-key", "some-region", "some-zone")
			err := credentialValidator.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the project id is missing", func() {
				credentialValidator = gcp.NewCredentialValidator("", "some-service-account-key", "some-region", "some-zone")
				Expect(credentialValidator.Validate()).To(MatchError("GCP project ID must be provided"))
			})

			It("returns an error when the service account key is missing", func() {
				credentialValidator = gcp.NewCredentialValidator("some-project-id", "", "some-region", "some-zone")
				Expect(credentialValidator.Validate()).To(MatchError("GCP service account key must be provided"))
			})

			It("returns an error when the region is missing", func() {
				credentialValidator = gcp.NewCredentialValidator("some-project-id", "some-service-account-key", "", "some-zone")
				Expect(credentialValidator.Validate()).To(MatchError("GCP region must be provided"))
			})

			It("returns an error when the zone is missing", func() {
				credentialValidator = gcp.NewCredentialValidator("some-project-id", "some-service-account-key", "some-region", "")
				Expect(credentialValidator.Validate()).To(MatchError("GCP zone must be provided"))
			})
		})
	})
})
