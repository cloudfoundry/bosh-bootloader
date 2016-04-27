package manifests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
)

var _ = Describe("ReleaseManifestBuilder", func() {
	var releaseManifestBuilder manifests.ReleaseManifestBuilder

	BeforeEach(func() {
		releaseManifestBuilder = manifests.NewReleaseManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all releases for manifest", func() {
			releases := releaseManifestBuilder.Build()

			Expect(releases).To(HaveLen(2))
			Expect(releases).To(ConsistOf([]manifests.Release{
				{
					Name: "bosh",
					URL:  "https://bosh.io/d/github.com/cloudfoundry/bosh?v=255.6",
					SHA1: "c80989984c4ec4c171f9d880c9f69586dade6389",
				},
				{
					Name: "bosh-aws-cpi",
					URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=48",
					SHA1: "2abfa1bed326238861e247a10674acf4f7ac48b8",
				},
			}))
		})
	})
})
