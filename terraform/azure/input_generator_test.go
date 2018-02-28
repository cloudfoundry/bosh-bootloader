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
				ClientID:       "client-id",
				ClientSecret:   "client-secret",
				Region:         "region",
				SubscriptionID: "subscription-id",
				TenantID:       "tenant-id",
			},
		}

		inputGenerator = azure.NewInputGenerator()
	})

	Context("Generate", func() {
		It("receives BBL state and returns a map of terraform variables", func() {
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputs).To(Equal(map[string]interface{}{
				"simple_env_id": "envid",
				"env_id":        state.EnvID,
				"region":        state.Azure.Region,
			}))
		})

		Context("given a long environment id", func() {
			It("shortens the id for simple_env_id", func() {
				state.EnvID = "super-long-environment-id-with-999"
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(inputs).To(Equal(map[string]interface{}{
					"simple_env_id": "superlongenvironment",
					"env_id":        state.EnvID,
					"region":        state.Azure.Region,
				}))
			})
		})

		Context("given a partial LB state", func() {
			It("does not generate input for the LB", func() {
				state.LB.Cert = "Cert content"
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(inputs)).To(Equal(3))
			})
		})

		Context("given an LB system domain", func() {
			It("returns system domain as input", func() {
				state.LB.Domain = "example.com"
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())
				Expect(inputs).To(HaveKeyWithValue("system_domain", "example.com"))
			})
		})

		Context("given a LB", func() {
			BeforeEach(func() {
				state.LB.Cert = "Cert content"
				state.LB.Key = "PFX password"
				state.LB.Domain = "example.com"
			})

			It("returns the expected inputs for the LB", func() {
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(inputs).To(Equal(map[string]interface{}{
					"simple_env_id":   "envid",
					"env_id":          state.EnvID,
					"region":          state.Azure.Region,
					"pfx_cert_base64": "Cert content",
					"pfx_password":    "PFX password",
					"system_domain":   "example.com",
				}))
			})
		})
	})

	Context("Credentials", func() {
		It("returns azure credentials", func() {
			credentials := inputGenerator.Credentials(state)

			Expect(credentials).To(Equal(map[string]string{
				"client_id":       "client-id",
				"client_secret":   "client-secret",
				"subscription_id": "subscription-id",
				"tenant_id":       "tenant-id",
			}))
		})
	})
})
