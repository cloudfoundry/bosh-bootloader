package helpers_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	compute "google.golang.org/api/compute/v1"
)

var _ = Describe("EnvIDManager", func() {
	var (
		envIDGenerator        *fakes.EnvIDGenerator
		gcpClientProvider     *fakes.GCPClientProvider
		gcpClient             *fakes.GCPClient
		infrastructureManager *fakes.InfrastructureManager
		envIDManager          helpers.EnvIDManager
	)

	BeforeEach(func() {
		envIDGenerator = &fakes.EnvIDGenerator{}
		envIDGenerator.GenerateCall.Returns.EnvID = "some-env-id"

		gcpClientProvider = &fakes.GCPClientProvider{}
		gcpClient = &fakes.GCPClient{}
		gcpClientProvider.ClientCall.Returns.Client = gcpClient

		infrastructureManager = &fakes.InfrastructureManager{}

		envIDManager = helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider, infrastructureManager)
	})

	Describe("Sync", func() {
		Context("when no previous env id exists", func() {
			It("calls env id generator if name is not passed in", func() {
				state, err := envIDManager.Sync(storage.State{}, "")
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDGenerator.GenerateCall.CallCount).To(Equal(1))
				Expect(state.EnvID).To(Equal("some-env-id"))
			})

			It("uses the name passed in if an environment does not exist", func() {
				state, err := envIDManager.Sync(storage.State{}, "some-other-env-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDGenerator.GenerateCall.CallCount).To(Equal(0))
				Expect(state.EnvID).To(Equal("some-other-env-id"))
			})

			Context("for gcp", func() {
				It("fails if a name of a pre-existing environment is passed in", func() {
					gcpClient.GetNetworksCall.Returns.NetworkList = &compute.NetworkList{
						Items: []*compute.Network{
							&compute.Network{},
						},
					}
					_, err := envIDManager.Sync(storage.State{
						IAAS: "gcp",
					}, "existing")

					Expect(gcpClient.GetNetworksCall.CallCount).To(Equal(1))
					Expect(gcpClient.GetNetworksCall.Receives.Name).To(Equal("existing-network"))

					Expect(err).To(MatchError("It looks like a bbl environment already exists with the name 'existing'. Please provide a different name."))
				})
			})

			Context("for aws", func() {
				It("fails if a name of a pre-existing environment is passed in", func() {
					infrastructureManager.ExistsCall.Returns.Exists = true
					_, err := envIDManager.Sync(storage.State{
						IAAS: "aws",
					}, "existing")

					Expect(infrastructureManager.ExistsCall.CallCount).To(Equal(1))
					Expect(infrastructureManager.ExistsCall.Receives.StackName).To(Equal("stack-existing"))

					Expect(err).To(MatchError("It looks like a bbl environment already exists with the name 'existing'. Please provide a different name."))
				})
			})
		})

		Context("when an env id exists in the state", func() {
			It("returns the existing env id", func() {
				state, err := envIDManager.Sync(storage.State{EnvID: "some-previous-env-id"}, "")
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDGenerator.GenerateCall.CallCount).To(Equal(0))
				Expect(state.EnvID).To(Equal("some-previous-env-id"))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the gcpClient cannot get networks", func() {
				gcpClient.GetNetworksCall.Returns.Error = errors.New("failed to get network list")

				_, err := envIDManager.Sync(storage.State{
					IAAS: "gcp",
				}, "existing")

				Expect(err).To(MatchError("failed to get network list"))
			})

			It("returns an error when the infrastructure manager cannot verify stack existence", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("failed to check stack existence")

				_, err := envIDManager.Sync(storage.State{
					IAAS: "aws",
				}, "existing")

				Expect(err).To(MatchError("failed to check stack existence"))
			})

			It("returns an error with a helpful message when an invalid name is provided", func() {
				_, err := envIDManager.Sync(storage.State{}, "some_bad_name")

				Expect(err).To(MatchError("Names must start with a letter and be alphanumeric or hyphenated."))
			})

			It("returns an error when the env id generator fails", func() {
				envIDGenerator.GenerateCall.Returns.Error = errors.New("failed to generate")
				_, err := envIDManager.Sync(storage.State{}, "")

				Expect(err).To(MatchError("failed to generate"))
			})

			It("returns an error when regex match string fails", func() {
				helpers.SetMatchString(func(string, string) (bool, error) {
					return false, errors.New("failed to match string")
				})

				_, err := envIDManager.Sync(storage.State{}, "some-name")

				Expect(err).To(MatchError("failed to match string"))

				helpers.ResetMatchString()
			})
		})
	})
})
