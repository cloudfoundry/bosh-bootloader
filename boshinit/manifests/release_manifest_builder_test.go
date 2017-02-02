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
			releases := releaseManifestBuilder.Build("some-bosh-url", "some-bosh-sha1", "bosh-some-cpi", "some-bosh-some-cpi-url", "some-bosh-some-cpi-sha1")

			Expect(releases).To(HaveLen(2))
			Expect(releases).To(ConsistOf([]manifests.Release{
				{
					Name: "bosh",
					URL:  "some-bosh-url",
					SHA1: "some-bosh-sha1",
				},
				{
					Name: "bosh-some-cpi",
					URL:  "some-bosh-some-cpi-url",
					SHA1: "some-bosh-some-cpi-sha1",
				},
			}))
		})
	})
})
