package gcp_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/application/gcp"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvironmentValidator", func() {
	var (
		boshClientProvider *fakes.BOSHClientProvider
		boshClient         *fakes.BOSHClient

		state                storage.State
		environmentValidator gcp.EnvironmentValidator
	)

	BeforeEach(func() {
		boshClientProvider = &fakes.BOSHClientProvider{}
		boshClient = &fakes.BOSHClient{}

		boshClientProvider.ClientCall.Returns.Client = boshClient

		environmentValidator = gcp.NewEnvironmentValidator(boshClientProvider)
	})

	It("returns an error if bosh client info fails", func() {
		boshClient.InfoCall.Returns.Error = errors.New("failed to talk to bosh")

		err := environmentValidator.Validate(storage.State{
			IAAS: "gcp",
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		})

		Expect(boshClientProvider.ClientCall.CallCount).To(Equal(1))
		Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
		Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
		Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))
		Expect(boshClient.InfoCall.CallCount).To(Equal(1))

		Expect(err).To(MatchError(application.BBLNotFound))
	})

	Context("when there is no director", func() {
		BeforeEach(func() {
			state = storage.State{
				IAAS:       "gcp",
				NoDirector: true,
			}
		})

		It("returns no error", func() {
			err := environmentValidator.Validate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.InfoCall.CallCount).To(Equal(0))
		})
	})
})
