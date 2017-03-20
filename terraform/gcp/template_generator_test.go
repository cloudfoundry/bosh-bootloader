package gcp_test

import (
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateGenerator", func() {

	var (
		zones             *fakes.Zones
		templateGenerator gcp.TemplateGenerator

		expectedTemplate []byte
	)

	BeforeEach(func() {
		zones = &fakes.Zones{}
		templateGenerator = gcp.NewTemplateGenerator(zones)

		zones.GetCall.Returns.Zones = []string{"z1", "z2", "z3"}
	})

	Describe("GenerateBackendService", func() {
		BeforeEach(func() {
			var err error
			expectedTemplate, err = ioutil.ReadFile("fixtures/backend_service.tf")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a backend service terraform template", func() {
			template := templateGenerator.GenerateBackendService("some-region")

			Expect(zones.GetCall.Receives.Region).To(Equal("some-region"))
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
			template := templateGenerator.GenerateInstanceGroups("some-region")

			Expect(zones.GetCall.Receives.Region).To(Equal("some-region"))
			Expect(template).To(Equal(string(expectedTemplate)))
		})
	})
})
