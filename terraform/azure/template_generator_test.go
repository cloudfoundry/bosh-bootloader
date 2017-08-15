package azure_test

import (
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateGenerator", func() {
	var (
		templateGenerator azure.TemplateGenerator
	)

	BeforeEach(func() {
		templateGenerator = azure.NewTemplateGenerator()
	})

	Describe("Generate", func() {
		It("generates a terraform template for azure", func() {
			expectedTemplate, err := ioutil.ReadFile("fixtures/azure_template_rg.tf")
			Expect(err).NotTo(HaveOccurred())

			template := templateGenerator.Generate(storage.State{
				EnvID: "azure-environment",
				Azure: storage.Azure{
					SubscriptionID: "subscription-id",
					TenantID:       "tenant-id",
					ClientID:       "client-id",
					ClientSecret:   "client-secret",
				},
			})
			Expect(template).To(Equal(string(expectedTemplate)))
		})
	})
})
