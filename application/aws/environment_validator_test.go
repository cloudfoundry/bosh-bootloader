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
		infrastructureManager *fakes.InfrastructureManager
		boshClientProvider    *fakes.BOSHClientProvider
		boshClient            *fakes.BOSHClient

		state                storage.State
		environmentValidator aws.EnvironmentValidator
	)

	BeforeEach(func() {
		infrastructureManager = &fakes.InfrastructureManager{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		boshClient = &fakes.BOSHClient{}

		infrastructureManager.ExistsCall.Returns.Exists = true
		boshClientProvider.ClientCall.Returns.Client = boshClient

		environmentValidator = aws.NewEnvironmentValidator(infrastructureManager, boshClientProvider)
	})

	It("returns a helpful error message when tf state and stack name are empty", func() {
		err := environmentValidator.Validate(storage.State{})

		Expect(infrastructureManager.ExistsCall.CallCount).To(Equal(0))
		Expect(err).To(MatchError(application.BBLNotFound))
	})

	Context("when there is no director", func() {
		BeforeEach(func() {
			state = storage.State{
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
				NoDirector: true,
			}
		})

		It("returns no error", func() {
			err := environmentValidator.Validate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(infrastructureManager.ExistsCall.CallCount).To(Equal(1))
			Expect(boshClient.InfoCall.CallCount).To(Equal(0))
		})
	})

	Context("when cloudformation was used to create infrastructure", func() {
		BeforeEach(func() {
			state = storage.State{
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
				BOSH: storage.BOSH{
					DirectorAddress:  "some-director-address",
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}
		})

		It("returns a helpful error message when stack does not exist", func() {
			infrastructureManager.ExistsCall.Returns.Exists = false

			err := environmentValidator.Validate(state)

			Expect(infrastructureManager.ExistsCall.CallCount).To(Equal(1))
			Expect(infrastructureManager.ExistsCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(err).To(MatchError(application.BBLNotFound))
		})

		It("returns a helpful error message when bosh does not exist", func() {
			boshClient.InfoCall.Returns.Error = errors.New("bosh is not available")

			err := environmentValidator.Validate(state)

			Expect(boshClientProvider.ClientCall.CallCount).To(Equal(1))
			Expect(boshClient.InfoCall.CallCount).To(Equal(1))
			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))
			Expect(err).To(MatchError(fmt.Sprintf("%s %s", application.DirectorNotReachable, "bosh is not available")))
		})

		Context("failure cases", func() {
			It("returns an error if Exists call on InfrastructureManager fails", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("exists call failed")
				err := environmentValidator.Validate(state)
				Expect(err).To(MatchError("exists call failed"))
			})
		})
	})

	Context("when terraform was used to create infrastruture", func() {
		var (
			state storage.State
		)

		BeforeEach(func() {
			state = storage.State{
				TFState: "some-tf-state",
				BOSH: storage.BOSH{
					DirectorAddress:  "some-director-address",
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}
		})

		It("returns a helpful error message when bosh does not exist", func() {
			boshClient.InfoCall.Returns.Error = errors.New("bosh is not available")

			err := environmentValidator.Validate(state)

			Expect(infrastructureManager.ExistsCall.CallCount).To(Equal(0))
			Expect(boshClientProvider.ClientCall.CallCount).To(Equal(1))
			Expect(boshClient.InfoCall.CallCount).To(Equal(1))
			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))
			Expect(err).To(MatchError(fmt.Sprintf("%s %s", application.DirectorNotReachable, "bosh is not available")))
		})
	})
})
