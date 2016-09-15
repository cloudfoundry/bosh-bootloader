package manifests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
)

var _ = Describe("DiskPoolsManifestBuilder", func() {
	var diskPoolsManifestBuilder manifests.DiskPoolsManifestBuilder

	BeforeEach(func() {
		diskPoolsManifestBuilder = manifests.NewDiskPoolsManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all disk pools for manifest", func() {
			diskPools := diskPoolsManifestBuilder.Build()

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
	})
})
