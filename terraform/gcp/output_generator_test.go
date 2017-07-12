package gcp_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	Describe("Generate", func() {
		var (
			executor        *fakes.TerraformExecutor
			outputGenerator gcp.OutputGenerator
		)

		BeforeEach(func() {
			executor = &fakes.TerraformExecutor{}
			outputGenerator = gcp.NewOutputGenerator(executor)

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

		Context("when a domain is provided", func() {
			It("formats the raw terraform output", func() {
				executor.OutputsCall.Returns.Outputs = map[string]interface{}{
					"system_domain_dns_servers": []interface{}{"domain-1", "domain-2", "domain-3"},
				}

				outputs, err := outputGenerator.Generate("")
				Expect(err).NotTo(HaveOccurred())
				Expect(outputs).To(HaveKeyWithValue("system_domain_dns_servers", []string{"domain-1", "domain-2", "domain-3"}))
			})
		})
	})
})
