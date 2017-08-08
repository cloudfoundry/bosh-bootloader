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
		azureUp commands.AzureUp

		azureClient *fakes.AzureClient
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		azureClient = &fakes.AzureClient{}

		azureUp = commands.NewAzureUp(azureClient, logger)
	})

	Describe("Execute", func() {
		It("validates credentials", func() {
			err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{
				Azure: storage.Azure{
					SubscriptionID: "subscription-id",
					TenantID:       "tenant-id",
					ClientID:       "client-id",
					ClientSecret:   "client-secret",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.CallCount).To(Equal(1))
			Expect(logger.StepCall.Messages).To(Equal([]string{"verifying credentials"}))

			Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))

			Expect(azureClient.ValidateCredentialsCall.Receives.TenantID).To(Equal("tenant-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.ClientID).To(Equal("client-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.ClientSecret).To(Equal("client-secret"))
		})

		Context("given invalid credentials", func() {
			BeforeEach(func() {
				azureClient.ValidateCredentialsCall.Returns.Error = errors.New("invalid credentials")
			})
			It("returns the error", func() {
				err := azureUp.Execute(commands.AzureUpConfig{}, storage.State{
					Azure: storage.Azure{
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
					},
				})
				Expect(err).To(MatchError("invalid credentials"))
				Expect(logger.StepCall.CallCount).To(Equal(1))
				Expect(logger.StepCall.Messages).To(Equal([]string{"verifying credentials"}))

				Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))

				Expect(azureClient.ValidateCredentialsCall.Receives.TenantID).To(Equal("tenant-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientID).To(Equal("client-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientSecret).To(Equal("client-secret"))
			})
		})
	})
})
