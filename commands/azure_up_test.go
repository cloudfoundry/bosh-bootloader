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

		azureClient      *fakes.AzureClient
		envIDManager     *fakes.EnvIDManager
		logger           *fakes.Logger
		stateStore       *fakes.StateStore
		terraformManager *fakes.TerraformManager
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		azureClient = &fakes.AzureClient{}
		envIDManager = &fakes.EnvIDManager{}
		stateStore = &fakes.StateStore{}
		terraformManager = &fakes.TerraformManager{}

		azureUp = commands.NewAzureUp(azureClient, logger, envIDManager, stateStore, terraformManager)
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

			Expect(azureClient.ValidateCredentialsCall.Receives.SubscriptionID).To(Equal("subscription-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.TenantID).To(Equal("tenant-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.ClientID).To(Equal("client-id"))
			Expect(azureClient.ValidateCredentialsCall.Receives.ClientSecret).To(Equal("client-secret"))

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
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
				Expect(err).To(MatchError("Error: credentials are invalid"))
				Expect(logger.StepCall.CallCount).To(Equal(1))
				Expect(logger.StepCall.Messages).To(Equal([]string{"verifying credentials"}))

				Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))

				Expect(azureClient.ValidateCredentialsCall.Receives.SubscriptionID).To(Equal("subscription-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.TenantID).To(Equal("tenant-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientID).To(Equal("client-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientSecret).To(Equal("client-secret"))
			})
		})

		Context("given an environment id", func() {
			var expectedEnvIDState storage.State

			BeforeEach(func() {
				expectedEnvIDState = storage.State{
					Azure: storage.Azure{
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
					},
				}
			})

			Context("when called with --name", func() {
				It("creates the environment", func() {
					err := azureUp.Execute(commands.AzureUpConfig{Name: "myenvid"}, expectedEnvIDState)
					Expect(err).NotTo(HaveOccurred())

					Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
					Expect(envIDManager.SyncCall.Receives.State).To(Equal(expectedEnvIDState))
					Expect(envIDManager.SyncCall.Receives.Name).To(Equal("myenvid"))
				})
			})

			Context("when callled without --name flag", func() {
				It("creates the environment", func() {
					err := azureUp.Execute(commands.AzureUpConfig{}, expectedEnvIDState)
					Expect(err).NotTo(HaveOccurred())

					Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
					Expect(envIDManager.SyncCall.Receives.State).To(Equal(expectedEnvIDState))
					Expect(envIDManager.SyncCall.Receives.Name).To(BeEmpty())
				})
			})
		})

		Context("when the no-director flag is provided", func() {
			BeforeEach(func() {
				terraformManager.ApplyCall.Returns.BBLState.NoDirector = true
			})

			It("does not create a bosh or update cloud config", func() {
				err := azureUp.Execute(commands.AzureUpConfig{
					NoDirector: true,
				}, storage.State{
					Azure: storage.Azure{
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.ApplyCall.CallCount).To(Equal(1))
				Expect(stateStore.SetCall.CallCount).To(Equal(2))
				Expect(stateStore.SetCall.Receives[0].State.NoDirector).To(Equal(true))
			})
		})
	})
})
