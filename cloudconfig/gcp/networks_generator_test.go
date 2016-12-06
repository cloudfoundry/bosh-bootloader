package gcp_test

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworksGenerator", func() {
	Describe("Generate", func() {
		It("returns a slice of networks for cloud config", func() {
			generator := gcp.NewNetworksGenerator(
				"some-network-name",
				"some-subnetwork-name",
				[]string{"some-tag", "some-other-tag"},
				[]string{"z1", "z2", "z3"},
			)

			networks, err := generator.Generate()
			Expect(err).NotTo(HaveOccurred())

			Expect(networks).To(ConsistOf(
				gcp.Network{
					Name: "private",
					Type: "manual",
					Subnets: []gcp.NetworkSubnet{
						{
							AZ:      "z1",
							Gateway: "10.0.16.1",
							Range:   "10.0.16.0/20",
							Reserved: []string{
								"10.0.16.2-10.0.16.3",
								"10.0.31.255",
							},
							Static: []string{
								"10.0.31.190-10.0.31.254",
							},
							CloudProperties: gcp.SubnetCloudProperties{
								EphemeralExternalIP: true,
								NetworkName:         "some-network-name",
								SubnetworkName:      "some-subnetwork-name",
								Tags: []string{
									"some-tag",
									"some-other-tag",
								},
							},
						},
						{
							AZ:      "z2",
							Gateway: "10.0.32.1",
							Range:   "10.0.32.0/20", Reserved: []string{
								"10.0.32.2-10.0.32.3",
								"10.0.47.255",
							},
							Static: []string{
								"10.0.47.190-10.0.47.254",
							},
							CloudProperties: gcp.SubnetCloudProperties{
								EphemeralExternalIP: true,
								NetworkName:         "some-network-name",
								SubnetworkName:      "some-subnetwork-name",
								Tags: []string{
									"some-tag",
									"some-other-tag",
								},
							},
						},
						{
							AZ:      "z3",
							Gateway: "10.0.48.1",
							Range:   "10.0.48.0/20",
							Reserved: []string{
								"10.0.48.2-10.0.48.3",
								"10.0.63.255",
							},
							Static: []string{
								"10.0.63.190-10.0.63.254",
							},
							CloudProperties: gcp.SubnetCloudProperties{
								EphemeralExternalIP: true,
								NetworkName:         "some-network-name",
								SubnetworkName:      "some-subnetwork-name",
								Tags: []string{
									"some-tag",
									"some-other-tag",
								},
							},
						},
					},
				},
			))
		})

		Context("failure cases", func() {
			It("returns an error when CIDR block cannot be parsed", func() {
				azs := []string{}
				for i := 0; i < 255; i++ {
					azs = append(azs, fmt.Sprintf("az%d", i))
				}
				generator := gcp.NewNetworksGenerator(
					"some-network-name",
					"some-subnetwork-name",
					[]string{"some-tag", "some-other-tag"},
					azs,
				)
				_, err := generator.Generate()
				Expect(err).To(MatchError(ContainSubstring("invalid ip")))
			})
		})
	})
})
