package bosh_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AZs Generator", func() {
	Describe("Generate", func() {
		It("returns cloud config AZs", func() {
			azsGenerator := bosh.NewAZsGenerator("us-east-1a", "us-east-1b", "us-east-1c")
			azs := azsGenerator.Generate()
			Expect(azs).To(ConsistOf(
				bosh.AZ{
					Name: "z1",
					CloudProperties: bosh.AZCloudProperties{
						AvailabilityZone: "us-east-1a",
					},
				},
				bosh.AZ{
					Name: "z2",
					CloudProperties: bosh.AZCloudProperties{
						AvailabilityZone: "us-east-1b",
					},
				},
				bosh.AZ{
					Name: "z3",
					CloudProperties: bosh.AZCloudProperties{
						AvailabilityZone: "us-east-1c",
					},
				}))
		})
	})
})
