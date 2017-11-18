package vsphere_test

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/vsphere"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InputGenerator", func() {
	var (
		inputGenerator vsphere.InputGenerator
	)

	It("receives state and returns a map of terraform variables", func() {
		inputs, err := inputGenerator.Generate(storage.State{
			VSphere: storage.VSphere{
				Subnet: "10.0.0.0/24",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(inputs).To(Equal(map[string]interface{}{
			"vsphere_subnet": "10.0.0.0/24",
			"external_ip":    "10.0.0.5",
		}))
	})
})
