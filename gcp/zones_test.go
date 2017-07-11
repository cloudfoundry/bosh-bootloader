package gcp_test

import (
	"github.com/cloudfoundry/bosh-bootloader/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("zones", func() {
	var zones gcp.Zones

	BeforeEach(func() {
		zones = gcp.NewZones()
	})

	Describe("get", func() {
		DescribeTable("returns a list of zones for a given region", func(region string, expectedZones []string) {
			actualZones := zones.Get(region)
			Expect(actualZones).To(Equal(expectedZones))
		},
			Entry("for us-west", "us-west1", []string{"us-west1-a", "us-west1-b", "us-west1-c"}),
			Entry("for us-central1", "us-central1", []string{"us-central1-a", "us-central1-b", "us-central1-c", "us-central1-f"}),
			Entry("for us-east1", "us-east1", []string{"us-east1-b", "us-east1-c", "us-east1-d"}),
			Entry("for europe-west1", "europe-west1", []string{"europe-west1-b", "europe-west1-c", "europe-west1-d"}),
			Entry("for asia-east1", "asia-east1", []string{"asia-east1-a", "asia-east1-b", "asia-east1-c"}),
			Entry("for asia-northeast1", "asia-northeast1", []string{"asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c"}),
		)
	})
})
