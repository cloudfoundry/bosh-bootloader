package manifests_test

import (
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SharedPropertiesManifestBuilder", func() {
	var sharedPropertiesManifestBuilder *manifests.SharedPropertiesManifestBuilder

	BeforeEach(func() {
		sharedPropertiesManifestBuilder = manifests.NewSharedPropertiesManifestBuilder()
	})

	Describe("AWS", func() {
		It("returns job properties for AWS", func() {
			aws := sharedPropertiesManifestBuilder.AWS(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
				SecurityGroup:   "some-security-group",
			})

			Expect(aws).To(Equal(manifests.AWSProperties{
				AccessKeyId:           "some-access-key-id",
				SecretAccessKey:       "some-secret-access-key",
				DefaultKeyName:        "some-key-name",
				DefaultSecurityGroups: []string{"some-security-group"},
				Region:                "some-region",
			}))
		})
	})

	Describe("Google", func() {
		It("returns job properties for Google", func() {
			gcp := sharedPropertiesManifestBuilder.Google(manifests.ManifestProperties{
				GCP: manifests.ManifestPropertiesGCP{
					Project: "some-project",
					JsonKey: `{"key":"value"}`,
				},
			})

			Expect(gcp).To(Equal(manifests.GoogleProperties{
				Project: "some-project",
				JsonKey: `{"key":"value"}`,
			}))
		})
	})
})
