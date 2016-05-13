package bosh_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiskTypesGenerator", func() {
	Describe("Generate", func() {
		It("returns a slice of disk types for cloud config", func() {
			generator := bosh.NewDiskTypesGenerator()
			diskTypes := generator.Generate()

			Expect(diskTypes).To(ConsistOf(
				bosh.DiskType{
					Name:     "default",
					DiskSize: 1024,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
			))
		})
	})
})
