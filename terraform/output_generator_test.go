package terraform_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	var (
		executor         *fakes.TerraformExecutor
		outputGenerator  terraform.OutputGenerator
		terraformOutputs map[string]interface{}
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		terraformOutputs = map[string]interface{}{
			"some-key": "some-value",
		}
		executor.OutputsCall.Returns.Outputs = terraformOutputs

		outputGenerator = terraform.NewOutputGenerator(executor)
	})

	It("returns an object to access all terraform outputs", func() {
		outputs, err := outputGenerator.Generate("some-key: some-value")
		Expect(err).NotTo(HaveOccurred())

		Expect(executor.OutputsCall.Receives.TFState).To(Equal("some-key: some-value"))
		Expect(outputs.Map).To(Equal(terraformOutputs))
	})

	Context("when executor outputs returns an error", func() {
		It("returns an empty map and the error", func() {
			executor.OutputsCall.Returns.Error = errors.New("executor outputs failed")

			outputs, err := outputGenerator.Generate("")
			Expect(err).To(MatchError("executor outputs failed"))
			Expect(outputs.Map).To(BeEmpty())
		})
	})
})
