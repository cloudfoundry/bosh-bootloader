package manifests_test

import (
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResourcePoolsManifestBuilder", func() {
	var resourcePoolsManifestBuilder manifests.ResourcePoolsManifestBuilder

	BeforeEach(func() {
		resourcePoolsManifestBuilder = manifests.NewResourcePoolsManifestBuilder()
	})

	Describe("ResourcePools", func() {
		It("returns all resource pools for manifest", func() {
			resourcePools := resourcePoolsManifestBuilder.Build(manifests.ManifestProperties{AvailabilityZone: "some-az"}, "some-stemcell-url", "some-stemcell-sha1")

			Expect(resourcePools).To(HaveLen(1))
			Expect(resourcePools).To(ConsistOf([]manifests.ResourcePool{
				{
					Name:    "vms",
					Network: "private",
					Stemcell: manifests.Stemcell{
						URL:  "some-stemcell-url",
						SHA1: "some-stemcell-sha1",
					},
					CloudProperties: manifests.ResourcePoolCloudProperties{
						InstanceType: "m3.xlarge",
						EphemeralDisk: manifests.EphemeralDisk{
							Size: 25000,
							Type: "gp2",
						},
						AvailabilityZone: "some-az",
					},
				},
			}))
		})
	})
})
