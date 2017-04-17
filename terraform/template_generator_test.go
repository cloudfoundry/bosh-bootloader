package terraform_test

import (
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateGenerator", func() {
	Describe("Generate", func() {
		var (
			gcpTemplateGenerator *fakes.TemplateGenerator
			awsTemplateGenerator *fakes.TemplateGenerator

			templateGenerator terraform.TemplateGenerator
		)

		BeforeEach(func() {
			gcpTemplateGenerator = &fakes.TemplateGenerator{}
			awsTemplateGenerator = &fakes.TemplateGenerator{}

			gcpTemplateGenerator.GenerateCall.Returns.Template = "some-gcp-template"
			awsTemplateGenerator.GenerateCall.Returns.Template = "some-aws-template"

			templateGenerator = terraform.NewTemplateGenerator(gcpTemplateGenerator, awsTemplateGenerator)
		})

		Context("when iaas is gcp", func() {
			It("returns the template from the gcp template generator", func() {
				template := templateGenerator.Generate(storage.State{
					IAAS: "gcp",
				})

				Expect(template).To(Equal("some-gcp-template"))
				Expect(gcpTemplateGenerator.GenerateCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
				}))
				Expect(awsTemplateGenerator.GenerateCall.CallCount).To(Equal(0))
			})
		})

		Context("when iaas is aws", func() {
			It("returns the template from the aws template generator", func() {
				template := templateGenerator.Generate(storage.State{
					IAAS: "aws",
				})

				Expect(template).To(Equal("some-aws-template"))
				Expect(gcpTemplateGenerator.GenerateCall.CallCount).To(Equal(0))
				Expect(awsTemplateGenerator.GenerateCall.Receives.State).To(Equal(storage.State{
					IAAS: "aws",
				}))
			})
		})

		Context("when iaas is invalid", func() {
			It("returns an empty string", func() {
				template := templateGenerator.Generate(storage.State{})

				Expect(template).To(Equal(""))
				Expect(gcpTemplateGenerator.GenerateCall.CallCount).To(Equal(0))
				Expect(awsTemplateGenerator.GenerateCall.CallCount).To(Equal(0))
			})
		})
	})
})
