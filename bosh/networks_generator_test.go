package bosh_test

import (
	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

	. "github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("optional network fields", func() {
	Describe("Network", func() {
		Context("when Subnets or CloudProperties are missing", func() {
			It("does not appear in the YAML output", func() {
				subnet := bosh.Network{
					Name: "some-network",
					Type: "some-type",
				}
				bytes, _ := candiedyaml.Marshal(subnet)
				Expect(bytes).To(MatchYAML(`{ "name": "some-network", "type": "some-type" }`))
			})
		})
		Context("when CloudProperties are present", func() {
			It("marshals them", func() {
				subnet := bosh.Network{
					Name: "some-network",
					Type: "dynamic",
					CloudProperties: &bosh.NetworkCloudProperties{
						Subnet: "some-subnet",
					},
				}
				bytes, _ := candiedyaml.Marshal(subnet)
				Expect(bytes).To(MatchYAML(`
name: some-network
type: dynamic
cloud_properties: { "subnet": "some-subnet" }`))
			})
		})
	})
	Describe("NetworkSubnetCloudProperties", func() {
		Context("when SecurityGroups are missing", func() {
			It("does not appear in the YAML output", func() {
				subnetCloudProps := bosh.SubnetCloudProperties{
					Subnet: "some-subnet",
				}
				bytes, _ := candiedyaml.Marshal(subnetCloudProps)
				Expect(bytes).To(MatchYAML(`{ "subnet": "some-subnet" }`))
			})
		})
	})
})

var _ = Describe("NetworksGenerator", func() {
	Describe("Generate", func() {
		It("returns a slice of networks for cloud config", func() {
			generator := bosh.NewNetworksGenerator([]bosh.SubnetInput{
				bosh.SubnetInput{
					AZ:             "us-east-1a",
					CIDR:           "10.0.16.0/20",
					Subnet:         "some-subnet-1",
					SecurityGroups: []string{"some-security-group-1"},
				},
				bosh.SubnetInput{
					AZ:             "us-east-1b",
					CIDR:           "10.0.32.0/20",
					Subnet:         "some-subnet-2",
					SecurityGroups: []string{"some-security-group-2"},
				},
				bosh.SubnetInput{
					AZ:             "us-east-1c",
					CIDR:           "10.0.48.0/20",
					Subnet:         "some-subnet-3",
					SecurityGroups: []string{"some-security-group-3"},
				},
			}, map[string]string{
				"us-east-1a": "z1",
				"us-east-1b": "z2",
				"us-east-1c": "z3",
			})
			networks, err := generator.Generate()
			Expect(err).NotTo(HaveOccurred())

			Expect(networks).To(ConsistOf(
				bosh.Network{
					Name: "private",
					Type: "manual",
					Subnets: []bosh.NetworkSubnet{
						{
							AZ:      "z1",
							Gateway: "10.0.16.1",
							Range:   "10.0.16.0/20",
							Reserved: []string{
								"10.0.16.2-10.0.16.3",
								"10.0.31.255",
							},
							CloudProperties: bosh.SubnetCloudProperties{
								Subnet: "some-subnet-1",
								SecurityGroups: []string{
									"some-security-group-1",
								},
							},
						},
						{
							AZ:      "z2",
							Gateway: "10.0.32.1",
							Range:   "10.0.32.0/20",
							Reserved: []string{
								"10.0.32.2-10.0.32.3",
								"10.0.47.255",
							},
							CloudProperties: bosh.SubnetCloudProperties{
								Subnet: "some-subnet-2",
								SecurityGroups: []string{
									"some-security-group-2",
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
							CloudProperties: bosh.SubnetCloudProperties{
								Subnet: "some-subnet-3",
								SecurityGroups: []string{
									"some-security-group-3",
								},
							},
						},
					},
				},
			))
		})
		Context("failure cases", func() {
			It("returns an error when CIDR block cannot be parsed", func() {

				generator := bosh.NewNetworksGenerator([]bosh.SubnetInput{
					bosh.SubnetInput{
						AZ:             "us-east-1a",
						CIDR:           "not-a-cidr-block",
						Subnet:         "some-subnet-1",
						SecurityGroups: []string{"some-security-group-1"},
					},
				}, map[string]string{
					"us-east-1a": "z1",
				})
				_, err := generator.Generate()
				Expect(err).To(MatchError(ContainSubstring("cannot parse CIDR block")))
			})

			It("returns an error when CIDR block is too small to contain the required reserved ips", func() {

				generator := bosh.NewNetworksGenerator([]bosh.SubnetInput{
					bosh.SubnetInput{
						AZ:             "us-east-1a",
						CIDR:           "10.0.16.0/32",
						Subnet:         "some-subnet-1",
						SecurityGroups: []string{"some-security-group-1"},
					},
				}, map[string]string{
					"us-east-1a": "z1",
				})
				_, err := generator.Generate()
				Expect(err).To(MatchError(ContainSubstring("not enough IPs allocated in CIDR block for subnet")))
			})
		})
	})
})
