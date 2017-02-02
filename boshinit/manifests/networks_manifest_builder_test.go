package manifests_test

import (
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworksManifestBuilder", func() {
	var networksManifestBuilder manifests.NetworksManifestBuilder

	BeforeEach(func() {
		networksManifestBuilder = manifests.NewNetworksManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all networks for manifest", func() {
			networks := networksManifestBuilder.Build(manifests.ManifestProperties{
				AWS: manifests.ManifestPropertiesAWS{SubnetID: "subnet-12345"}})

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

		It("returns networks with aws cloud properties", func() {
			networks := networksManifestBuilder.Build(manifests.ManifestProperties{
				AWS: manifests.ManifestPropertiesAWS{SubnetID: "subnet-12345"}})
			Expect(networks[0].Subnets[0].CloudProperties).To(Equal(manifests.NetworksCloudProperties{
				Subnet: "subnet-12345",
			}))
		})

		It("returns networks with gcp cloud properties", func() {
			networks := networksManifestBuilder.Build(manifests.ManifestProperties{
				GCP: manifests.ManifestPropertiesGCP{
					NetworkName:    "some-network",
					SubnetworkName: "some-subnet",
					BOSHTag:        "some-bosh-open-tag",
					InternalTag:    "some-internal-tag",
				},
			})

			ip := false
			Expect(networks[0].Subnets[0].CloudProperties).To(Equal(manifests.NetworksCloudProperties{
				NetworkName:         "some-network",
				SubnetworkName:      "some-subnet",
				EphemeralExternalIP: &ip,
				Tags: []string{
					"some-bosh-open-tag",
					"some-internal-tag",
				},
			}))
		})
	})
})
