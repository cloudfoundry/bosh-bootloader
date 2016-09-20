package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

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
					Name:     "1GB",
					DiskSize: 1024,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
				bosh.DiskType{
					Name:     "5GB",
					DiskSize: 5120,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
				bosh.DiskType{
					Name:     "10GB",
					DiskSize: 10240,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
				bosh.DiskType{
					Name:     "50GB",
					DiskSize: 51200,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
				bosh.DiskType{
					Name:     "100GB",
					DiskSize: 102400,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
				bosh.DiskType{
					Name:     "500GB",
					DiskSize: 512000,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
				bosh.DiskType{
					Name:     "1TB",
					DiskSize: 1048576,
					CloudProperties: bosh.DiskTypeCloudProperties{
						Type:      "gp2",
						Encrypted: true,
					},
				},
			))
		})
	})
})
