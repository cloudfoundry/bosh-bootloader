package gcp_test

import (
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateGenerator", func() {
	var (
		templateGenerator gcp.TemplateGenerator
		expectedTemplate  []byte
		zones             []string
	)

	BeforeEach(func() {
		templateGenerator = gcp.NewTemplateGenerator()
		zones = []string{"z1", "z2", "z3"}
	})

	Describe("Generate", func() {
		DescribeTable("generates a terraform template for gcp", func(fixtureFilename, region, lbType, domain string) {
			expectedTemplate, err := ioutil.ReadFile(fixtureFilename)
			Expect(err).NotTo(HaveOccurred())

			template := templateGenerator.Generate(storage.State{
				GCP: storage.GCP{
					Region: region,
					Zones:  zones,
				},
				LB: storage.LB{
					Type:   lbType,
					Domain: domain,
				},
			})
			Expect(template).To(Equal(string(expectedTemplate)))
		},
			Entry("when no lb type is provided", "fixtures/gcp_template_no_lb.tf", "some-region", "", ""),
			Entry("when a concourse lb type is provided", "fixtures/gcp_template_concourse_lb.tf", "some-region", "concourse", ""),
			Entry("when a cf lb type is provided", "fixtures/gcp_template_cf_lb.tf", "some-region", "cf", ""),
			Entry("when a cf lb type is provided with a domain", "fixtures/gcp_template_cf_lb_dns.tf", "some-region", "cf", "some-domain"),
		)
	})

	Describe("GenerateBackendService", func() {
		BeforeEach(func() {
			var err error
			expectedTemplate, err = ioutil.ReadFile("fixtures/backend_service.tf")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a backend service terraform template", func() {
			template := templateGenerator.GenerateBackendService(zones)

			Expect(template).To(Equal(string(expectedTemplate)))
		})
	})

	Describe("GenerateInstanceGroups", func() {
		BeforeEach(func() {
			var err error
			expectedTemplate, err = ioutil.ReadFile("fixtures/instance_groups.tf")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a backend service terraform template", func() {
			template := templateGenerator.GenerateInstanceGroups(zones)

			Expect(template).To(Equal(string(expectedTemplate)))
		})
	})

	Describe("GenerateSubnetCidrs", func() {
		BeforeEach(func() {
			var err error
			expectedTemplate, err = ioutil.ReadFile("fixtures/subnet_cidrs.tf")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a backend service terraform template", func() {
			template := templateGenerator.GenerateSubnetCidrs(zones)

			Expect(template).To(Equal(string(expectedTemplate)))
		})
	})
})
