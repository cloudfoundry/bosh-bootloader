package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
)

var _ = Describe("CloudProviderManifestBuilder", func() {
	var cloudProviderManifestBuilder boshinit.CloudProviderManifestBuilder

	BeforeEach(func() {
		cloudProviderManifestBuilder = boshinit.NewCloudProviderManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all cloud provider fields for manifest", func() {
			cloudProvider := cloudProviderManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})

			Expect(cloudProvider).To(Equal(boshinit.CloudProvider{
				Template: boshinit.Template{
					Name:    "aws_cpi",
					Release: "bosh-aws-cpi",
				},

				SSHTunnel: boshinit.SSHTunnel{
					Host:       "some-elastic-ip",
					Port:       22,
					User:       "vcap",
					PrivateKey: "./bosh.pem",
				},

				MBus: "https://mbus:mbus-password@some-elastic-ip:6868",

				Properties: boshinit.CloudProviderProperties{
					AWS: boshinit.AWSProperties{
						AccessKeyId:           "some-access-key-id",
						SecretAccessKey:       "some-secret-access-key",
						DefaultKeyName:        "some-key-name",
						DefaultSecurityGroups: []string{"bosh"},
						Region:                "some-region",
					},

					Agent: boshinit.AgentProperties{
						MBus: "https://mbus:mbus-password@0.0.0.0:6868",
					},

					Blobstore: boshinit.BlobstoreProperties{
						Provider: "local",
						Path:     "/var/vcap/micro_bosh/data/cache",
					},

					NTP: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
				},
			}))
		})
	})
})
