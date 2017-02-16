package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VMTypes Generator", func() {
	Describe("Generate", func() {
		It("returns cloud config VM Types", func() {
			vmTypesGenerator := bosh.NewVMTypesGenerator()
			vmTypes := vmTypesGenerator.Generate()
			ExpectToContainVMType(vmTypes, "m3.medium", "m3.medium")
			ExpectToContainVMType(vmTypes, "m3.large", "m3.large")
			ExpectToContainVMType(vmTypes, "m3.xlarge", "m3.xlarge")
			ExpectToContainVMType(vmTypes, "m3.2xlarge", "m3.2xlarge")
			ExpectToContainVMType(vmTypes, "m4.large", "m4.large")
			ExpectToContainVMType(vmTypes, "m4.xlarge", "m4.xlarge")
			ExpectToContainVMType(vmTypes, "m4.2xlarge", "m4.2xlarge")
			ExpectToContainVMType(vmTypes, "m4.4xlarge", "m4.4xlarge")
			ExpectToContainVMType(vmTypes, "m4.10xlarge", "m4.10xlarge")
			ExpectToContainVMType(vmTypes, "c3.large", "c3.large")
			ExpectToContainVMType(vmTypes, "c3.xlarge", "c3.xlarge")
			ExpectToContainVMType(vmTypes, "c3.2xlarge", "c3.2xlarge")
			ExpectToContainVMType(vmTypes, "c3.4xlarge", "c3.4xlarge")
			ExpectToContainVMType(vmTypes, "c3.8xlarge", "c3.8xlarge")
			ExpectToContainVMType(vmTypes, "c4.large", "c4.large")
			ExpectToContainVMType(vmTypes, "c4.xlarge", "c4.xlarge")
			ExpectToContainVMType(vmTypes, "c4.2xlarge", "c4.2xlarge")
			ExpectToContainVMType(vmTypes, "c4.4xlarge", "c4.4xlarge")
			ExpectToContainVMType(vmTypes, "c4.8xlarge", "c4.8xlarge")
			ExpectToContainVMType(vmTypes, "r3.large", "r3.large")
			ExpectToContainVMType(vmTypes, "r3.xlarge", "r3.xlarge")
			ExpectToContainVMType(vmTypes, "r3.2xlarge", "r3.2xlarge")
			ExpectToContainVMType(vmTypes, "r3.4xlarge", "r3.4xlarge")
			ExpectToContainVMType(vmTypes, "r3.8xlarge", "r3.8xlarge")
			ExpectToContainVMType(vmTypes, "t2.nano", "t2.nano")
			ExpectToContainVMType(vmTypes, "t2.micro", "t2.micro")
			ExpectToContainVMType(vmTypes, "t2.small", "t2.small")
			ExpectToContainVMType(vmTypes, "t2.medium", "t2.medium")
			ExpectToContainVMType(vmTypes, "t2.large", "t2.large")
			ExpectToContainVMType(vmTypes, "default", "m3.medium")
			ExpectToContainVMType(vmTypes, "extra-small", "t2.small")
			ExpectToContainVMType(vmTypes, "small", "m3.large")
			ExpectToContainVMType(vmTypes, "medium", "m4.xlarge")
			ExpectToContainVMType(vmTypes, "large", "m4.2xlarge")
			ExpectToContainVMType(vmTypes, "extra-large", "m4.4xlarge")
		})
	})
})

func ExpectToContainVMType(vmTypes []bosh.VMType, vmName, vmType string) {
	Expect(vmTypes).To(ContainElement(
		bosh.VMType{
			Name: vmName,
			CloudProperties: &bosh.VMTypeCloudProperties{
				InstanceType: vmType,
				EphemeralDisk: &bosh.EphemeralDisk{
					Size: 10240,
					Type: "gp2",
				},
			},
		}))
}
