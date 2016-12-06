package gcp_test

import (
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AZs Generator", func() {
	Describe("Generate", func() {
		It("returns cloud config AZs", func() {
			azsGenerator := gcp.NewAZsGenerator("us-east1-a", "us-east1-b", "us-east1-c")
			azs := azsGenerator.Generate()
			Expect(azs).To(ConsistOf(
				gcp.AZ{
					Name: "z1",
					CloudProperties: gcp.AZCloudProperties{
						Zone: "us-east1-a",
					},
				},
				gcp.AZ{
					Name: "z2",
					CloudProperties: gcp.AZCloudProperties{
						Zone: "us-east1-b",
					},
				},
				gcp.AZ{
					Name: "z3",
					CloudProperties: gcp.AZCloudProperties{
						Zone: "us-east1-c",
					},
				}))
		})
	})
})
