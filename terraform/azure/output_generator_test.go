package azure_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform/azure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	Describe("Generate", func() {
		var (
			executor        *fakes.TerraformExecutor
			outputGenerator azure.OutputGenerator
		)

		BeforeEach(func() {
			executor = &fakes.TerraformExecutor{}
			outputGenerator = azure.NewOutputGenerator(executor)

			executor.OutputsCall.Returns.Outputs = map[string]interface{}{
				"some-key": "some-value",
			}
		})

		It("returns the outputs from the terraform state", func() {
			outputs, err := outputGenerator.Generate("some-key: some-value")
			Expect(err).NotTo(HaveOccurred())
			Expect(outputs).To(HaveKeyWithValue("some-key", "some-value"))
		})

		Context("when executor outputs returns an error", func() {
			It("returns an empty map and the error", func() {
				executor.OutputsCall.Returns.Error = errors.New("executor outputs failed")

				outputs, err := outputGenerator.Generate("")
				Expect(err).To(MatchError("executor outputs failed"))
				Expect(outputs).To(BeEmpty())
			})
		})
	})
})
