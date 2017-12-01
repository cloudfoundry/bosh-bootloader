package azure_test

import (
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
		DescribeTable("generates a terraform template for azure", func(fixture, lbType, domain string) {
			expectedTemplate, err := ioutil.ReadFile(fixture)
			Expect(err).NotTo(HaveOccurred())

			template := templateGenerator.Generate(storage.State{
				EnvID: "azure-environment",
				Azure: storage.Azure{
					SubscriptionID: "subscription-id",
					TenantID:       "tenant-id",
					Region:         "my-location",
					ClientID:       "client-id",
					ClientSecret:   "client-secret",
				},
				LB: storage.LB{
					Type:   lbType,
					Domain: domain,
				},
			})
			Expect(template).To(Equal(string(expectedTemplate)))
		},
			Entry("when no lb type is provided", "fixtures/base.tf", "", ""),
			Entry("when a cf lb type is provided with a domain", "fixtures/cf_lb.tf", "cf", "some-domain"),
		)
	})
})
