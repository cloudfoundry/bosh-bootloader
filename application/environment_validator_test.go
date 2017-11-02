package application_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvironmentValidator", func() {
	Describe("Validate", func() {
		var (
			boshClientProvider *fakes.BOSHClientProvider
			boshClient         *fakes.BOSHClient

			environmentValidator application.EnvironmentValidator
		)

		BeforeEach(func() {
			boshClientProvider = &fakes.BOSHClientProvider{}
			boshClient = &fakes.BOSHClient{}
			boshClientProvider.ClientCall.Returns.Client = boshClient

			environmentValidator = application.NewEnvironmentValidator(boshClientProvider)
		})

		Context("when the director is unavailable", func() {
			BeforeEach(func() {
				boshClient.InfoCall.Returns.Error = errors.New("bosh is not available")
			})

			It("returns a helpful error message", func() {
				err := environmentValidator.Validate(storage.State{
					BOSH: storage.BOSH{
						DirectorAddress:  "some-director-address",
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
					},
				})

				Expect(boshClientProvider.ClientCall.CallCount).To(Equal(1))
				Expect(boshClient.InfoCall.CallCount).To(Equal(1))
				Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))
				Expect(err).To(MatchError(fmt.Sprintf("%s %s", application.DirectorNotReachable, "bosh is not available")))
			})
		})

		Context("on GCP", func() {
			It("checks availability zones", func() {
				err := environmentValidator.Validate(storage.State{
					IAAS:       "gcp",
					NoDirector: true,
					GCP: storage.GCP{
						Zones: []string{"zone-1", "zone-2"},
					},
				})

				Expect(err).NotTo(HaveOccurred())
			})

			Context("when zones are missing from the state", func() {
				It("returns an error", func() {
					err := environmentValidator.Validate(storage.State{
						IAAS:       "gcp",
						NoDirector: true,
					})

					Expect(err).To(MatchError("bbl state is missing availability zones; have you run bbl up?"))
				})
			})
		})
	})
})
