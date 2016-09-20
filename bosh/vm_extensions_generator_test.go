package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VMExtensionsGenerator", func() {
	Describe("Generate", func() {
		It("returns cloud config vm extensions", func() {
			input := []bosh.LoadBalancerExtension{
				{
					Name:    "lb",
					ELBName: "some-lb",
				},
				{
					Name:    "another-lb",
					ELBName: "some-other-lb",
					SecurityGroups: []string{
						"some-security-group",
						"some-other-security-group",
					},
				},
			}

			vmExtensions := bosh.NewVMExtensionsGenerator(input).Generate()

			Expect(vmExtensions).To(HaveLen(8))
			Expect(vmExtensions).To(Equal([]bosh.VMExtension{
				{
					Name: "5GB_ephemeral_disk",
					CloudProperties: bosh.VMExtensionCloudProperties{
						EphemeralDisk: &bosh.VMExtensionEphemeralDisk{
							Size: 5120,
							Type: "gp2",
						},
					},
				},
				{
					Name: "10GB_ephemeral_disk",
					CloudProperties: bosh.VMExtensionCloudProperties{
						EphemeralDisk: &bosh.VMExtensionEphemeralDisk{
							Size: 10240,
							Type: "gp2",
						},
					},
				},
				{
					Name: "50GB_ephemeral_disk",
					CloudProperties: bosh.VMExtensionCloudProperties{
						EphemeralDisk: &bosh.VMExtensionEphemeralDisk{
							Size: 51200,
							Type: "gp2",
						},
					},
				},
				{
					Name: "100GB_ephemeral_disk",
					CloudProperties: bosh.VMExtensionCloudProperties{
						EphemeralDisk: &bosh.VMExtensionEphemeralDisk{
							Size: 102400,
							Type: "gp2",
						},
					},
				},
				{
					Name: "500GB_ephemeral_disk",
					CloudProperties: bosh.VMExtensionCloudProperties{
						EphemeralDisk: &bosh.VMExtensionEphemeralDisk{
							Size: 512000,
							Type: "gp2",
						},
					},
				},
				{
					Name: "1TB_ephemeral_disk",
					CloudProperties: bosh.VMExtensionCloudProperties{
						EphemeralDisk: &bosh.VMExtensionEphemeralDisk{
							Size: 1048576,
							Type: "gp2",
						},
					},
				},
				{
					Name: "lb",
					CloudProperties: bosh.VMExtensionCloudProperties{
						ELBS: []string{"some-lb"},
					},
				},
				{
					Name: "another-lb",
					CloudProperties: bosh.VMExtensionCloudProperties{
						ELBS:           []string{"some-other-lb"},
						SecurityGroups: []string{"some-security-group", "some-other-security-group"},
					},
				},
			}))
		})
	})
})
