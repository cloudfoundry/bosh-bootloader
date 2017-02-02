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
		It("returns all resource pools for manifest for aws", func() {
			resourcePools := resourcePoolsManifestBuilder.Build("aws", manifests.ManifestProperties{
				AWS: manifests.ManifestPropertiesAWS{
					AvailabilityZone: "some-az",
				},
			}, "some-stemcell-url", "some-stemcell-sha1")

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

		It("returns all resource pools for manifest for gcp", func() {
			resourcePools := resourcePoolsManifestBuilder.Build("gcp", manifests.ManifestProperties{
				GCP: manifests.ManifestPropertiesGCP{
					Zone: "some-zone",
				},
			}, "some-stemcell-url", "some-stemcell-sha1")

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
						Zone:           "some-zone",
						MachineType:    "n1-standard-4",
						RootDiskSizeGB: 25,
						RootDiskType:   "pd-standard",
						ServiceScopes: []string{
							"compute",
							"devstorage.full_control",
						},
					},
				},
			}))
		})
	})
})
