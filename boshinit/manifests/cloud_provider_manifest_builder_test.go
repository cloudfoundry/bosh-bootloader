package manifests_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
)

var _ = Describe("CloudProviderManifestBuilder", func() {
	var (
		cloudProviderManifestBuilder manifests.CloudProviderManifestBuilder
		stringGenerator              *fakes.StringGenerator
	)

	BeforeEach(func() {
		stringGenerator = &fakes.StringGenerator{}
		cloudProviderManifestBuilder = manifests.NewCloudProviderManifestBuilder(stringGenerator)
		stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
			return fmt.Sprintf("%s%s", prefix, "some-random-string"), nil
		}
	})

	Describe("Build", func() {
		It("returns all cloud provider fields for manifest", func() {
			cloudProvider, _, err := cloudProviderManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
				SecurityGroup:   "some-security-group",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudProvider).To(Equal(manifests.CloudProvider{
				Template: manifests.Template{
					Name:    "aws_cpi",
					Release: "bosh-aws-cpi",
				},

				SSHTunnel: manifests.SSHTunnel{
					Host:       "some-elastic-ip",
					Port:       22,
					User:       "vcap",
					PrivateKey: "./bosh.pem",
				},

				MBus: "https://mbus-user-some-random-string:mbus-some-random-string@some-elastic-ip:6868",

				Properties: manifests.CloudProviderProperties{
					AWS: manifests.AWSProperties{
						AccessKeyId:           "some-access-key-id",
						SecretAccessKey:       "some-secret-access-key",
						DefaultKeyName:        "some-key-name",
						DefaultSecurityGroups: []string{"some-security-group"},
						Region:                "some-region",
					},

					Agent: manifests.AgentProperties{
						MBus: "https://mbus-user-some-random-string:mbus-some-random-string@0.0.0.0:6868",
					},

					Blobstore: manifests.BlobstoreProperties{
						Provider: "local",
						Path:     "/var/vcap/micro_bosh/data/cache",
					},
				},
			}))

			Expect(stringGenerator.GenerateCall.Receives.Prefixes).To(Equal([]string{"mbus-user-", "mbus-"}))
			Expect(stringGenerator.GenerateCall.Receives.Lengths).To(Equal([]int{15, 15}))
		})

		It("returns manifest properties with new credentials", func() {
			_, manifestProperties, err := cloudProviderManifestBuilder.Build(manifests.ManifestProperties{})
			Expect(err).NotTo(HaveOccurred())

			Expect(manifestProperties.Credentials.MBusUsername).To(Equal("mbus-user-some-random-string"))
			Expect(manifestProperties.Credentials.MBusPassword).To(Equal("mbus-some-random-string"))
		})

		It("returns manifest and manifest properties with existing credentials", func() {
			cloudProvider, manifestProperties, err := cloudProviderManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
				SecurityGroup:   "some-security-group",
				Credentials: manifests.InternalCredentials{
					MBusUsername: "some-persisted-mbus-username",
					MBusPassword: "some-persisted-mbus-password",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudProvider.MBus).To(Equal("https://some-persisted-mbus-username:some-persisted-mbus-password@some-elastic-ip:6868"))
			Expect(cloudProvider.Properties.Agent.MBus).To(Equal("https://some-persisted-mbus-username:some-persisted-mbus-password@0.0.0.0:6868"))
			Expect(manifestProperties.Credentials.MBusPassword).To(Equal("some-persisted-mbus-password"))
		})

		Context("when string generator cannot generated a string", func() {
			BeforeEach(func() {
				stringGenerator.GenerateCall.Stub = nil
				stringGenerator.GenerateCall.Returns.Error = fmt.Errorf("foo")
			})

			It("forwards the error", func() {
				_, _, err := cloudProviderManifestBuilder.Build(manifests.ManifestProperties{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("foo"))
			})
		})
	})
})
