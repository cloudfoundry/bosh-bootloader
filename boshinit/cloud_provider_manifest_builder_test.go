package boshinit_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
)

var _ = Describe("CloudProviderManifestBuilder", func() {
	var (
		cloudProviderManifestBuilder boshinit.CloudProviderManifestBuilder
		uuidGenerator                *fakes.UUIDGenerator
	)

	BeforeEach(func() {
		uuidGenerator = &fakes.UUIDGenerator{}
		cloudProviderManifestBuilder = boshinit.NewCloudProviderManifestBuilder(uuidGenerator)
	})

	Describe("Build", func() {
		BeforeEach(func() {
			uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{
				{String: "fake-randomly-generated-password"},
			}
		})

		It("returns all cloud provider fields for manifest", func() {
			cloudProvider, err := cloudProviderManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
				SecurityGroup:   "some-security-group",
			})
			Expect(err).NotTo(HaveOccurred())

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

				MBus: "https://mbus:fake-randomly-generated-password@some-elastic-ip:6868",

				Properties: boshinit.CloudProviderProperties{
					AWS: boshinit.AWSProperties{
						AccessKeyId:           "some-access-key-id",
						SecretAccessKey:       "some-secret-access-key",
						DefaultKeyName:        "some-key-name",
						DefaultSecurityGroups: []string{"some-security-group"},
						Region:                "some-region",
					},

					Agent: boshinit.AgentProperties{
						MBus: "https://mbus:fake-randomly-generated-password@0.0.0.0:6868",
					},

					Blobstore: boshinit.BlobstoreProperties{
						Provider: "local",
						Path:     "/var/vcap/micro_bosh/data/cache",
					},

					NTP: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
				},
			}))
		})

		Context("when uuidGenerator cannot generated a uuid", func() {
			BeforeEach(func() {
				uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{{Error: fmt.Errorf("foo")}}
			})

			It("forwards the error", func() {
				_, err := cloudProviderManifestBuilder.Build(boshinit.ManifestProperties{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("foo"))
			})
		})

		Context("when bbl is run a second time", func() {
			It("doesn't regenerate the message bus password", func() {

			})
		})
	})
})
