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
			ExpectToContainVMType(vmTypes, "m3.medium")
			ExpectToContainVMType(vmTypes, "m3.large")
			ExpectToContainVMType(vmTypes, "m3.xlarge")
			ExpectToContainVMType(vmTypes, "m3.2xlarge")
			ExpectToContainVMType(vmTypes, "m4.large")
			ExpectToContainVMType(vmTypes, "m4.xlarge")
			ExpectToContainVMType(vmTypes, "m4.2xlarge")
			ExpectToContainVMType(vmTypes, "m4.4xlarge")
			ExpectToContainVMType(vmTypes, "m4.10xlarge")
			ExpectToContainVMType(vmTypes, "c3.large")
			ExpectToContainVMType(vmTypes, "c3.xlarge")
			ExpectToContainVMType(vmTypes, "c3.2xlarge")
			ExpectToContainVMType(vmTypes, "c3.4xlarge")
			ExpectToContainVMType(vmTypes, "c3.8xlarge")
			ExpectToContainVMType(vmTypes, "c4.large")
			ExpectToContainVMType(vmTypes, "c4.xlarge")
			ExpectToContainVMType(vmTypes, "c4.2xlarge")
			ExpectToContainVMType(vmTypes, "c4.4xlarge")
			ExpectToContainVMType(vmTypes, "c4.8xlarge")
			ExpectToContainVMType(vmTypes, "r3.large")
			ExpectToContainVMType(vmTypes, "r3.xlarge")
			ExpectToContainVMType(vmTypes, "r3.2xlarge")
			ExpectToContainVMType(vmTypes, "r3.4xlarge")
			ExpectToContainVMType(vmTypes, "r3.8xlarge")
			ExpectToContainVMType(vmTypes, "t2.nano")
			ExpectToContainVMType(vmTypes, "t2.micro")
			ExpectToContainVMType(vmTypes, "t2.small")
			ExpectToContainVMType(vmTypes, "t2.medium")
			ExpectToContainVMType(vmTypes, "t2.large")
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
