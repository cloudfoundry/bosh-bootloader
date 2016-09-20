package manifests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
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
					URL:  "https://s3.amazonaws.com/bbl-precompiled-bosh-releases/release-bosh-257.9-on-ubuntu-trusty-stemcell-3262.12.tgz",
					SHA1: "68125b0e36f599c79f10f8809d328a6dea7e2cd3",
				},
				{
					Name: "bosh-aws-cpi",
					URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=60",
					SHA1: "8e40a9ff892204007889037f094a1b0d23777058",
				},
			}))
		})
	})
})
