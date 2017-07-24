package aws_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform/aws"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	var (
		executor        *fakes.TerraformExecutor
		outputGenerator aws.OutputGenerator
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		executor.OutputsCall.Returns.Outputs = map[string]interface{}{
			"some-key": "some-value",
		}

		outputGenerator = aws.NewOutputGenerator(executor)
	})

	It("returns all terraform outputs", func() {
		outputs, err := outputGenerator.Generate("some-key: some-value")
		Expect(err).NotTo(HaveOccurred())

		Expect(executor.OutputsCall.Receives.TFState).To(Equal("some-key: some-value"))
		Expect(outputs).To(HaveKeyWithValue("some-key", "some-value"))
	})

	Context("when a domain is provided", func() {
		It("formats the raw terraform output", func() {
			executor.OutputsCall.Returns.Outputs = map[string]interface{}{
				"env_dns_zone_name_servers": []interface{}{"domain-1", "domain-2", "domain-3"},
			}

			outputs, err := outputGenerator.Generate("")
			Expect(err).NotTo(HaveOccurred())
			Expect(outputs).To(HaveKeyWithValue("env_dns_zone_name_servers", []string{"domain-1", "domain-2", "domain-3"}))
		})
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
