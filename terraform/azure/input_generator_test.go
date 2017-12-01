package azure_test

import (
	"io/ioutil"
	"os"
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
				Region:       "region",
				SubscriptionID: "subscription-id",
				TenantID:       "tenant-id",
			},
		}

		inputGenerator = azure.NewInputGenerator()
	})

	It("receives BBL state and returns a map of terraform variables", func() {
		inputs, err := inputGenerator.Generate(state)
		Expect(err).NotTo(HaveOccurred())

		Expect(inputs).To(Equal(map[string]interface{}{
			"simple_env_id":   "envid",
			"env_id":          state.EnvID,
			"region":        state.Azure.Region,
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

			Expect(inputs).To(Equal(map[string]interface{}{
				"simple_env_id":   "superlongenvironment",
				"env_id":          state.EnvID,
				"region":        state.Azure.Region,
				"subscription_id": state.Azure.SubscriptionID,
				"tenant_id":       state.Azure.TenantID,
				"client_id":       state.Azure.ClientID,
				"client_secret":   state.Azure.ClientSecret,
			}))
		})
	})

	Context("given a partial LB state", func() {
		It("does not generate input for the LB", func(){
			state.LB.Cert = "certpath"
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(inputs)).To(Equal(7))
		})
	})

	Context("given a LB", func() {
		var (
			cert *os.File
			key *os.File
		)

		BeforeEach(func() {
			var err error

			cert, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			cert.WriteString("cert content")

			key, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			key.WriteString("secret key")

			state.LB.Cert = cert.Name()
			state.LB.Key = key.Name()
		})

		AfterEach(func() {
			cert.Close()
			key.Close()
		})

		It("returns the expected inputs for the LB", func() {
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())

			inputs["pfx_cert_base64"] = ""

			Expect(inputs).To(Equal(map[string]interface{}{
				"simple_env_id":   "envid",
				"env_id":          state.EnvID,
				"region":        state.Azure.Region,
				"subscription_id": state.Azure.SubscriptionID,
				"tenant_id":       state.Azure.TenantID,
				"client_id":       state.Azure.ClientID,
				"client_secret":   state.Azure.ClientSecret,
				"pfx_cert_base64": "",
				"pfx_key":         "secret key",
			}))
		})

		It("converts the cert to base64", func() {
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(inputs["pfx_cert_base64"]).To(Equal("Y2VydCBjb250ZW50"))
		})
	})
})
