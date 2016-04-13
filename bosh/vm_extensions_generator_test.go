package bosh_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VMExtensionsGenerator", func() {
	Describe("Generate", func() {
		It("returns cloud config vm extensions", func() {
			vmExtensions := bosh.NewVMExtensionsGenerator("some-lb").Generate()

			Expect(vmExtensions).To(HaveLen(1))
			Expect(vmExtensions[0]).To(Equal(bosh.VMExtension{
				Name: "lb",
				CloudProperties: bosh.VMExtensionCloudProperties{
					ELBS: []string{"some-lb"},
				},
			}))
		})
	})
})
