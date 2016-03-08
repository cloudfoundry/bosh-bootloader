package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
)

var _ = Describe("ResourcePoolsManifestBuilder", func() {
	var resourcePoolsManifestBuilder boshinit.ResourcePoolsManifestBuilder

	BeforeEach(func() {
		resourcePoolsManifestBuilder = boshinit.NewResourcePoolsManifestBuilder()
	})

	Describe("ResourcePools", func() {
		It("returns all resource pools for manifest", func() {
			resourcePools := resourcePoolsManifestBuilder.Build("some-az")

			Expect(resourcePools).To(HaveLen(1))
			Expect(resourcePools).To(ConsistOf([]boshinit.ResourcePool{
				{
					Name:    "vms",
					Network: "private",
					Stemcell: boshinit.Stemcell{
						URL:  "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3012",
						SHA1: "3380b55948abe4c437dee97f67d2d8df4eec3fc1",
					},
					CloudProperties: boshinit.ResourcePoolCloudProperties{
						InstanceType: "m3.xlarge",
						EphemeralDisk: boshinit.EphemeralDisk{
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
