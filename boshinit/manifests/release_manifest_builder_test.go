package manifests_test

import (
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseManifestBuilder", func() {
	var releaseManifestBuilder manifests.ReleaseManifestBuilder

	BeforeEach(func() {
		releaseManifestBuilder = manifests.NewReleaseManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all releases for manifest", func() {
			releases := releaseManifestBuilder.Build("some-bosh-url", "some-bosh-sha1", "some-bosh-aws-cpi-url", "some-bosh-aws-cpi-sha1")

			Expect(releases).To(HaveLen(2))
			Expect(releases).To(ConsistOf([]manifests.Release{
				{
					Name: "bosh",
					URL:  "some-bosh-url",
					SHA1: "some-bosh-sha1",
				},
				{
					Name: "bosh-aws-cpi",
					URL:  "some-bosh-aws-cpi-url",
					SHA1: "some-bosh-aws-cpi-sha1",
				},
			}))
		})
	})
})
