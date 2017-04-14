package terraform_test

import (
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	Describe("Generate", func() {
		var (
			gcpOutputGenerator *fakes.OutputGenerator

			outputGenerator terraform.OutputGenerator
		)

		BeforeEach(func() {
			gcpOutputGenerator = &fakes.OutputGenerator{}
			gcpOutputGenerator.GenerateCall.Returns.Outputs = map[string]interface{}{
				"some-output": "some-value",
			}

			outputGenerator = terraform.NewOutputGenerator(gcpOutputGenerator)
		})

		Context("when iaas is gcp", func() {
			It("returns the outputs from the gcp output generator", func() {
				output, err := outputGenerator.Generate(storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(output).To(Equal(map[string]interface{}{
					"some-output": "some-value",
				}))
				Expect(gcpOutputGenerator.GenerateCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
				}))
			})
		})

		Context("failure cases", func() {
			Context("when iaas is invalid", func() {
				It("returns an error", func() {
					_, err := outputGenerator.Generate(storage.State{
						IAAS: "some-invalid-iaas",
					})
					Expect(err).To(MatchError(`invalid iaas: "some-invalid-iaas"`))

					Expect(gcpOutputGenerator.GenerateCall.CallCount).To(Equal(0))
				})
			})
		})
	})
})
