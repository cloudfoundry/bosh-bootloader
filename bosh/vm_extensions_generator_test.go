package bosh_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

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
				},
			}

			vmExtensions := bosh.NewVMExtensionsGenerator(input).Generate()

			Expect(vmExtensions).To(HaveLen(2))
			Expect(vmExtensions).To(Equal([]bosh.VMExtension{
				{
					Name: "lb",
					CloudProperties: bosh.VMExtensionCloudProperties{
						ELBS: []string{"some-lb"},
					},
				},
				{
					Name: "another-lb",
					CloudProperties: bosh.VMExtensionCloudProperties{
						ELBS: []string{"some-other-lb"},
					},
				},
			}))
		})
	})
})
