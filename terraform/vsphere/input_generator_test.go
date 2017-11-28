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
				Subnet:          "10.0.0.0/24",
				Cluster:         "the-cluster",
				Network:         "the-network",
				VCenterUser:     "the-user",
				VCenterPassword: "the-password",
				VCenterIP:       "the-ip",
				VCenterDC:       "the-datacenter",
				VCenterRP:       "the-resource-pool",
				VCenterDS:       "the-datastore",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(inputs).To(Equal(map[string]interface{}{
			"vsphere_subnet":            "10.0.0.0/24",
			"jumpbox_ip":                "10.0.0.5",
			"bosh_director_internal_ip": "10.0.0.6",
			"internal_gw":               "10.0.0.1",
			"network_name":              "the-network",
			"vcenter_cluster":           "the-cluster",
			"vcenter_user":              "the-user",
			"vcenter_password":          "the-password",
			"vcenter_ip":                "the-ip",
			"vcenter_dc":                "the-datacenter",
			"vcenter_rp":                "the-resource-pool",
			"vcenter_ds":                "the-datastore",
		}))
	})
})
