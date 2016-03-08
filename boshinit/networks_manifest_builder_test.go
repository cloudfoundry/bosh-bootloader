package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
)

var _ = Describe("NetworksManifestBuilder", func() {
	var networksManifestBuilder boshinit.NetworksManifestBuilder

	BeforeEach(func() {
		networksManifestBuilder = boshinit.NewNetworksManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all networks for manifest", func() {
			networks := networksManifestBuilder.Build("subnet-12345")

			Expect(networks).To(HaveLen(2))
			Expect(networks).To(ConsistOf([]boshinit.Network{
				{
					Name: "private",
					Type: "manual",
					Subnets: []boshinit.Subnet{
						{
							Range:   "10.0.0.0/24",
							Gateway: "10.0.0.1",
							DNS:     []string{"10.0.0.2"},
							CloudProperties: boshinit.NetworksCloudProperties{
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
