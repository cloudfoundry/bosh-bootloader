package vsphere_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/vsphere"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpsGenerator", func() {
	Describe("GenerateVars", func() {
		var (
			opsGenerator     vsphere.OpsGenerator
			terraformManager *fakes.TerraformManager
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}
			opsGenerator = vsphere.NewOpsGenerator(terraformManager)

			terraformManager.GetOutputsCall.Returns.Outputs.Map = map[string]interface{}{
				"some-key": "some-value",
			}
		})

		It("generates the cloud-config vars", func() {
			vars, err := opsGenerator.GenerateVars(storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(vars).To(MatchYAML(`---
some-key: some-value
`))
		})

		Context("when terraform manager get outputs fails", func() {
			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("kiwi")
			})

			It("returns the error", func() {
				_, err := opsGenerator.GenerateVars(storage.State{})
				Expect(err).To(MatchError("Get terraform outputs: kiwi"))
			})
		})
	})
})
