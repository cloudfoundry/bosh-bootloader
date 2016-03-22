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
			ExpectToContainVMType(vmTypes, "c3.large")
			ExpectToContainVMType(vmTypes, "c3.xlarge")
			ExpectToContainVMType(vmTypes, "m3.large")
			ExpectToContainVMType(vmTypes, "m3.medium")
			ExpectToContainVMType(vmTypes, "r3.xlarge")
			ExpectToContainVMType(vmTypes, "c4.large")
			ExpectToContainVMType(vmTypes, "c3.2xlarge")
			ExpectToContainVMType(vmTypes, "t2.micro")
		})
	})
})

func ExpectToContainVMType(vmTypes []bosh.VMType, vmType string) {
	Expect(vmTypes).To(ContainElement(
		bosh.VMType{
			Name: vmType,
			CloudProperties: &bosh.VMTypeCloudProperties{
				InstanceType: vmType,
				EphemeralDisk: &bosh.EphemeralDisk{
					Size: 1024,
					Type: "gp2",
				},
			},
		}))
}
