package terraform_test

import (
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InputGenerator", func() {
	Describe("Generate", func() {
		var (
			gcpInputGenerator *fakes.InputGenerator
			awsInputGenerator *fakes.InputGenerator

			inputGenerator terraform.InputGenerator
		)

		BeforeEach(func() {
			gcpInputGenerator = &fakes.InputGenerator{}
			gcpInputGenerator.GenerateCall.Returns.Inputs = map[string]string{
				"some-input": "some-value",
			}
			awsInputGenerator = &fakes.InputGenerator{}
			awsInputGenerator.GenerateCall.Returns.Inputs = map[string]string{
				"some-input": "some-value",
			}

			inputGenerator = terraform.NewInputGenerator(gcpInputGenerator, awsInputGenerator)
		})

		Context("when iaas is gcp", func() {
			It("returns the inputs from the gcp input generator", func() {
				input, err := inputGenerator.Generate(storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(input).To(Equal(map[string]string{
					"some-input": "some-value",
				}))
				Expect(gcpInputGenerator.GenerateCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
				}))
				Expect(awsInputGenerator.GenerateCall.CallCount).To(Equal(0))
			})
		})

		Context("when iaas is aws", func() {
			It("returns the inputs from the aws input generator", func() {
				input, err := inputGenerator.Generate(storage.State{
					IAAS: "aws",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(input).To(Equal(map[string]string{
					"some-input": "some-value",
				}))
				Expect(gcpInputGenerator.GenerateCall.CallCount).To(Equal(0))
				Expect(awsInputGenerator.GenerateCall.Receives.State).To(Equal(storage.State{
					IAAS: "aws",
				}))
			})
		})

		Context("failure cases", func() {
			Context("when iaas is invalid", func() {
				It("returns an error", func() {
					_, err := inputGenerator.Generate(storage.State{
						IAAS: "some-invalid-iaas",
					})
					Expect(err).To(MatchError(`invalid iaas: "some-invalid-iaas"`))

					Expect(gcpInputGenerator.GenerateCall.CallCount).To(Equal(0))
					Expect(awsInputGenerator.GenerateCall.CallCount).To(Equal(0))
				})
			})
		})
	})
})
