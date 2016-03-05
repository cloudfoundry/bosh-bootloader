package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
)

var _ = Describe("ReleaseManifestBuilder", func() {
	var releaseManifestBuilder boshinit.ReleaseManifestBuilder

	BeforeEach(func() {
		releaseManifestBuilder = boshinit.NewReleaseManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all releases for manifest", func() {
			releases := releaseManifestBuilder.Build()

			Expect(releases).To(HaveLen(2))
			Expect(releases).To(ConsistOf([]boshinit.Release{
				{
					Name: "bosh",
					URL:  "https://bosh.io/d/github.com/cloudfoundry/bosh?v=255.3",
					SHA1: "1a3d61f968b9719d9afbd160a02930c464958bf4",
				},
				{
					Name: "bosh-aws-cpi",
					URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=44",
					SHA1: "a1fe03071e8b9bf1fa97a4022151081bf144c8bc",
				},
			}))
		})
	})
})
