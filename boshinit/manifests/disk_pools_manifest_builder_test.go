package manifests_test

import (
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiskPoolsManifestBuilder", func() {
	var diskPoolsManifestBuilder manifests.DiskPoolsManifestBuilder

	BeforeEach(func() {
		diskPoolsManifestBuilder = manifests.NewDiskPoolsManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all disk pools for manifest for aws", func() {
			diskPools := diskPoolsManifestBuilder.Build("aws")

			Expect(diskPools).To(HaveLen(1))
			Expect(diskPools).To(ConsistOf([]manifests.DiskPool{
				{
					Name:     "disks",
					DiskSize: 80 * 1024,
					CloudProperties: manifests.DiskPoolsCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
			}))
		})

		It("returns all disk pools for manifest for gcp", func() {
			diskPools := diskPoolsManifestBuilder.Build("gcp")

			Expect(diskPools).To(HaveLen(1))
			Expect(diskPools).To(ConsistOf([]manifests.DiskPool{
				{
					Name:     "disks",
					DiskSize: 80 * 1024,
					CloudProperties: manifests.DiskPoolsCloudProperties{
						Type:      "pd-standard",
						Encrypted: true,
					},
				},
			}))
		})
	})
})
