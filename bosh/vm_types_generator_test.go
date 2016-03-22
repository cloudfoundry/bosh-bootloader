package bosh_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VMTypes Generator", func() {
	Describe("Generate", func() {
		It("returns cloud config VM Types", func() {
			vmTypesGenerator := bosh.NewVMTypesGenerator()
			vmTypes := vmTypesGenerator.Generate()
			Expect(vmTypes).To(ConsistOf(
				bosh.VMType{
					Name: "default",
					CloudProperties: &bosh.VMTypeCloudProperties{
						InstanceType: "m3.medium",
						EphemeralDisk: &bosh.EphemeralDisk{
							Size: 1024,
							Type: "gp2",
						},
					},
				}))
		})
	})
})
