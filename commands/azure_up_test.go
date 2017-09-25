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

		azureClient        *fakes.AzureClient
		boshManager        *fakes.BOSHManager
		cloudConfigManager *fakes.CloudConfigManager
		envIDManager       *fakes.EnvIDManager
		stateStore         *fakes.StateStore
		terraformManager   *fakes.TerraformManager
	)

	BeforeEach(func() {
		azureClient = &fakes.AzureClient{}
		boshManager = &fakes.BOSHManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		envIDManager = &fakes.EnvIDManager{}
		stateStore = &fakes.StateStore{}
		terraformManager = &fakes.TerraformManager{}

		azureUp = commands.NewAzureUp(azureClient, boshManager, cloudConfigManager, envIDManager, stateStore, terraformManager)
	})

	Describe("Execute", func() {
		var state storage.State

		BeforeEach(func() {
			state = storage.State{
				Azure: storage.Azure{
					SubscriptionID: "subscription-id",
					TenantID:       "tenant-id",
					ClientID:       "client-id",
					ClientSecret:   "client-secret",
				},
			}
		})

		It("creates the environment", func() {
			err := azureUp.Execute(commands.UpConfig{}, state)
			Expect(err).NotTo(HaveOccurred())

			By("validating credentials", func() {
				Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))

				Expect(azureClient.ValidateCredentialsCall.Receives.SubscriptionID).To(Equal("subscription-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.TenantID).To(Equal("tenant-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientID).To(Equal("client-id"))
				Expect(azureClient.ValidateCredentialsCall.Receives.ClientSecret).To(Equal("client-secret"))
			})
		})

		Context("given invalid credentials", func() {
			BeforeEach(func() {
				azureClient.ValidateCredentialsCall.Returns.Error = errors.New("fig")
			})

			It("returns the error", func() {
				err := azureUp.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("Validate credentials: fig"))
				Expect(azureClient.ValidateCredentialsCall.CallCount).To(Equal(1))
			})
		})
	})
})
