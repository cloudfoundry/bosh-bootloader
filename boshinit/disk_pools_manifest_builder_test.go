package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
)

var _ = Describe("DiskPoolsManifestBuilder", func() {
	var diskPoolsManifestBuilder boshinit.DiskPoolsManifestBuilder

	BeforeEach(func() {
		diskPoolsManifestBuilder = boshinit.NewDiskPoolsManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all disk pools for manifest", func() {
			diskPools := diskPoolsManifestBuilder.Build()

			Expect(diskPools).To(HaveLen(1))
			Expect(diskPools).To(ConsistOf([]boshinit.DiskPool{
				{
					Name:     "disks",
					DiskSize: 20000,
					CloudProperties: boshinit.DiskPoolsCloudProperties{
						Type: "gp2",
					},
				},
			}))
		})
	})
})
