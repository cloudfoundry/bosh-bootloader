package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AzureUp", func() {
	var (
		azureUp     commands.AzureUp
		azureClient *fakes.AzureClient
	)

	BeforeEach(func() {
		azureClient = &fakes.AzureClient{}
		azureUp = commands.NewAzureUp(azureClient)
	})

	Describe("Execute", func() {
		var incomingState storage.State
		BeforeEach(func() {
			incomingState = storage.State{
				Azure: storage.Azure{
					SubscriptionID: "subscription-id",
					TenantID:       "tenant-id",
					ClientID:       "client-id",
					ClientSecret:   "client-secret",
				},
			}
		})

		It("creates the environment", func() {
			returnedState, err := azureUp.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))

			Expect(azureClient.ValidateCredentialsCall.Receives.SubscriptionID).To(Equal("subscription-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.TenantID).To(Equal("tenant-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.ClientID).To(Equal("client-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.ClientSecret).To(Equal("client-secret"))

			Expect(returnedState).To(Equal(incomingState))
		})

		Context("given invalid credentials", func() {
			BeforeEach(func() {
				azureClient.ValidateCredentialsCall.Returns.Error = errors.New("fig")
			})

			It("returns the error", func() {
				_, err := azureUp.Execute(storage.State{})
				Expect(err).To(MatchError("Validate credentials: fig"))
				Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))
			})
		})
	})
})
