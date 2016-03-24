package manifests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
)

var _ = Describe("NetworksManifestBuilder", func() {
	var networksManifestBuilder manifests.NetworksManifestBuilder

	BeforeEach(func() {
		networksManifestBuilder = manifests.NewNetworksManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all networks for manifest", func() {
			networks := networksManifestBuilder.Build(manifests.ManifestProperties{SubnetID: "subnet-12345"})

			Expect(networks).To(HaveLen(2))
			Expect(networks).To(ConsistOf([]manifests.Network{
				{
					Name: "private",
					Type: "manual",
					Subnets: []manifests.Subnet{
						{
							Range:   "10.0.0.0/24",
							Gateway: "10.0.0.1",
							DNS:     []string{"10.0.0.2"},
							CloudProperties: manifests.NetworksCloudProperties{
								Subnet: "subnet-12345",
							},
						},
					},
				},
				{
					Name: "public",
					Type: "vip",
				},
			}))
		})
	})
})
