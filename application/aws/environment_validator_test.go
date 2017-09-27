package aws_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/application/aws"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvironmentValidator", func() {
	var (
		boshClientProvider   *fakes.BOSHClientProvider
		boshClient           *fakes.BOSHClient
		environmentValidator aws.EnvironmentValidator
	)

	BeforeEach(func() {
		boshClientProvider = &fakes.BOSHClientProvider{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider.ClientCall.Returns.Client = boshClient

		environmentValidator = aws.NewEnvironmentValidator(boshClientProvider)
	})

	Context("when there is no director", func() {
		It("returns no error", func() {
			err := environmentValidator.Validate(storage.State{NoDirector: true})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.InfoCall.CallCount).To(Equal(0))
		})
	})

	Context("when the director is unavailable", func() {
		BeforeEach(func() {
			boshClient.InfoCall.Returns.Error = errors.New("bosh is not available")
		})

		It("returns a helpful error message", func() {
			err := environmentValidator.Validate(storage.State{
				TFState: "some-tf-state",
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
})
