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

	Describe("Generate", func() {
		It("receives state and returns a map of terraform variables", func() {
			inputs, err := inputGenerator.Generate(storage.State{
				EnvID: "banana",
				VSphere: storage.VSphere{
					SubnetCIDR:       "10.0.0.0/24",
					Network:          "the-network",
					VCenterCluster:   "the-cluster",
					VCenterUser:      "the-user",
					VCenterPassword:  "the-password",
					VCenterIP:        "the-ip",
					VCenterDC:        "the-datacenter",
					VCenterRP:        "the-resource-pool",
					VCenterDS:        "the-datastore",
					VCenterDisks:     "the-disks",
					VCenterTemplates: "the-templates",
					VCenterVMs:       "the-vms",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(inputs).To(Equal(map[string]interface{}{
				"env_id":               "banana",
				"vsphere_subnet_cidr":  "10.0.0.0/24",
				"jumpbox_ip":           "10.0.0.5",
				"director_internal_ip": "10.0.0.6",
				"internal_gw":          "10.0.0.1",
				"network_name":         "the-network",
				"vcenter_cluster":      "the-cluster",
				"vcenter_ip":           "the-ip",
				"vcenter_dc":           "the-datacenter",
				"vcenter_rp":           "the-resource-pool",
				"vcenter_ds":           "the-datastore",
				"vcenter_disks":        "the-disks",
				"vcenter_templates":    "the-templates",
				"vcenter_vms":          "the-vms",
			}))
		})
	})
	Describe("Credentials", func() {
		It("returns the vsphere credentials", func() {
			state := storage.State{
				VSphere: storage.VSphere{
					VCenterUser:     "the-user",
					VCenterPassword: "the-password",
				},
			}

			credentials := inputGenerator.Credentials(state)

			Expect(credentials).To(Equal(map[string]string{
				"vcenter_user":     "the-user",
				"vcenter_password": "the-password",
			}))
		})
	})
})
