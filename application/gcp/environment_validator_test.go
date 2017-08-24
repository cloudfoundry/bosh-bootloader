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

		environmentValidator gcp.EnvironmentValidator

		state storage.State
	)

	BeforeEach(func() {
		boshClient = &fakes.BOSHClient{}
		boshClientProvider = &fakes.BOSHClientProvider{}

		boshClientProvider.ClientCall.Returns.Client = boshClient

		environmentValidator = gcp.NewEnvironmentValidator(boshClientProvider)
	})

	Context("when there is no director", func() {
		BeforeEach(func() {
			state = storage.State{
				TFState:    "some-tf-state",
				NoDirector: true,
			}
		})

		It("returns no error", func() {
			err := environmentValidator.Validate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.InfoCall.CallCount).To(Equal(0))
		})
	})

	Context("when there is a director and terraform state", func() {
		BeforeEach(func() {
			state = storage.State{
				TFState: "some-tf-state",
				BOSH: storage.BOSH{
					DirectorAddress:  "some-address",
					DirectorUsername: "some-username",
					DirectorPassword: "some-password",
				},
			}
		})

		It("returns no errors when bosh director exists", func() {
			err := environmentValidator.Validate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.CallCount).To(Equal(1))
			Expect(boshClient.InfoCall.CallCount).To(Equal(1))
			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-password"))
		})

		Context("when the director cannot be reached", func() {
			It("returns a helpful error message", func() {
				boshClient.InfoCall.Returns.Error = errors.New("some error")
				err := environmentValidator.Validate(state)
				Expect(err).To(MatchError(application.DirectorNotReachable))
			})
		})
	})

	Context("when there is a terraform state", func() {
		It("returns no error", func() {
			err := environmentValidator.Validate(storage.State{
				IAAS:    "gcp",
				TFState: "tf-state",
			})

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when tf state is empty", func() {
		It("returns a BBLNotFound error when tf state is empty", func() {
			err := environmentValidator.Validate(storage.State{
				IAAS:    "gcp",
				TFState: "",
			})

			Expect(err).To(MatchError(application.BBLNotFound))
		})
	})
})
