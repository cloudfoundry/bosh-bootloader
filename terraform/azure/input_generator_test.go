package azure_test

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InputGenerator", func() {
	var (
		inputGenerator azure.InputGenerator

		state storage.State
	)

	BeforeEach(func() {
		state = storage.State{
			IAAS:  "azure",
			EnvID: "env-id",
			Azure: storage.Azure{
				SubscriptionID: "subscription-id",
				TenantID:       "tenant-id",
				ClientID:       "client-id",
				ClientSecret:   "client-secret",
			},
		}

		inputGenerator = azure.NewInputGenerator()
	})

	It("receives BBL state and returns a map of terraform variables", func() {
		inputs, err := inputGenerator.Generate(state)
		Expect(err).NotTo(HaveOccurred())

		Expect(inputs).To(Equal(map[string]string{
			"simple_env_id":   "envid",
			"env_id":          state.EnvID,
			"location":        "West US",
			"subscription_id": state.Azure.SubscriptionID,
			"tenant_id":       state.Azure.TenantID,
			"client_id":       state.Azure.ClientID,
			"client_secret":   state.Azure.ClientSecret,
		}))
	})

	Context("given a long environment id", func() {
		It("shortens the id for simple_env_id", func() {
			state.EnvID = "super-long-environment-id-with-999"
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputs).To(Equal(map[string]string{
				"simple_env_id":   "superlongenvironment",
				"env_id":          state.EnvID,
				"location":        "West US",
				"subscription_id": state.Azure.SubscriptionID,
				"tenant_id":       state.Azure.TenantID,
				"client_id":       state.Azure.ClientID,
				"client_secret":   state.Azure.ClientSecret,
			}))
		})
	})
})
