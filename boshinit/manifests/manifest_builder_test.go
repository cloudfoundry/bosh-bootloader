package manifests_test

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("ManifestBuilder", func() {
	var (
		logger                       *fakes.Logger
		sslKeyPairGenerator          *fakes.SSLKeyPairGenerator
		stringGenerator              *fakes.StringGenerator
		manifestBuilder              manifests.ManifestBuilder
		manifestProperties           manifests.ManifestProperties
		cloudProviderManifestBuilder manifests.CloudProviderManifestBuilder
		jobsManifestBuilder          manifests.JobsManifestBuilder
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		sslKeyPairGenerator = &fakes.SSLKeyPairGenerator{}
		stringGenerator = &fakes.StringGenerator{}
		cloudProviderManifestBuilder = manifests.NewCloudProviderManifestBuilder(stringGenerator)
		jobsManifestBuilder = manifests.NewJobsManifestBuilder(stringGenerator)

		manifestBuilder = manifests.NewManifestBuilder(logger, sslKeyPairGenerator, stringGenerator, cloudProviderManifestBuilder, jobsManifestBuilder)
		manifestProperties = manifests.ManifestProperties{
			DirectorName:     "bosh-name",
			DirectorUsername: "bosh-username",
			DirectorPassword: "bosh-password",
			SubnetID:         "subnet-12345",
			AvailabilityZone: "some-az",
			CACommonName:     "BOSH Bootloader",
			ElasticIP:        "52.0.112.12",
			AccessKeyID:      "some-access-key-id",
			SecretAccessKey:  "some-secret-access-key",
			DefaultKeyName:   "some-key-name",
			Region:           "some-region",
			SecurityGroup:    "some-security-group",
		}

		stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
			return fmt.Sprintf("%s%s", prefix, "some-random-string"), nil
		}
	})

	Describe("Build", func() {
		It("builds the bosh-init manifest and updates the manifest properties", func() {
			sslKeyPairGenerator.GenerateCall.Returns.KeyPair = ssl.KeyPair{
				CA:          []byte(ca),
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}

			manifest, manifestProperties, err := manifestBuilder.Build(manifestProperties)
			Expect(err).NotTo(HaveOccurred())

			expectedAWSProperties := manifests.AWSProperties{
				AccessKeyId:           "some-access-key-id",
				SecretAccessKey:       "some-secret-access-key",
				DefaultKeyName:        "some-key-name",
				DefaultSecurityGroups: []string{"some-security-group"},
				Region:                "some-region",
			}

			Expect(manifest.Name).To(Equal("bosh"))
			Expect(manifest.Releases[0].Name).To(Equal("bosh"))
			Expect(manifest.ResourcePools[0].CloudProperties.AvailabilityZone).To(Equal("some-az"))
			Expect(manifest.DiskPools[0].Name).To(Equal("disks"))
			Expect(manifest.Networks[0].Subnets[0].CloudProperties.Subnet).To(Equal("subnet-12345"))
			Expect(manifest.Jobs[0].Networks[1].StaticIPs[0]).To(Equal("52.0.112.12"))
			Expect(manifest.Jobs[0].Properties.AWS).To(Equal(expectedAWSProperties))
			Expect(manifest.Jobs[0].Properties.Director.Name).To(Equal("bosh-name"))
			Expect(manifest.Jobs[0].Properties.Director.SSL).To(Equal(manifests.SSLProperties{
				Cert: certificate,
				Key:  privateKey,
			}))
			Expect(manifest.CloudProvider.Properties.AWS).To(Equal(expectedAWSProperties))
			Expect(manifest.CloudProvider.SSHTunnel.Host).To(Equal("52.0.112.12"))
			Expect(manifest.CloudProvider.MBus).To(Equal("https://mbus-user-some-random-string:mbus-some-random-string@52.0.112.12:6868"))

			Expect(sslKeyPairGenerator.GenerateCall.Receives.CACommonName).To(Equal("BOSH Bootloader"))
			Expect(sslKeyPairGenerator.GenerateCall.Receives.CertCommonName).To(Equal("52.0.112.12"))
			Expect(sslKeyPairGenerator.GenerateCall.CallCount).To(Equal(1))

			Expect(manifestProperties).To(Equal(
				manifests.ManifestProperties{
					DirectorName:     "bosh-name",
					DirectorUsername: "bosh-username",
					DirectorPassword: "bosh-password",
					SubnetID:         "subnet-12345",
					AvailabilityZone: "some-az",
					CACommonName:     "BOSH Bootloader",
					ElasticIP:        "52.0.112.12",
					AccessKeyID:      "some-access-key-id",
					SecretAccessKey:  "some-secret-access-key",
					DefaultKeyName:   "some-key-name",
					Region:           "some-region",
					SecurityGroup:    "some-security-group",
					SSLKeyPair: ssl.KeyPair{
						CA:          []byte(ca),
						Certificate: []byte(certificate),
						PrivateKey:  []byte(privateKey),
					},
					Credentials: manifests.InternalCredentials{
						MBusUsername:              "mbus-user-some-random-string",
						NatsUsername:              "nats-user-some-random-string",
						PostgresUsername:          "postgres-user-some-random-string",
						RegistryUsername:          "registry-user-some-random-string",
						BlobstoreDirectorUsername: "blobstore-director-user-some-random-string",
						BlobstoreAgentUsername:    "blobstore-agent-user-some-random-string",
						HMUsername:                "hm-user-some-random-string",
						MBusPassword:              "mbus-some-random-string",
						NatsPassword:              "nats-some-random-string",
						PostgresPassword:          "postgres-some-random-string",
						RegistryPassword:          "registry-some-random-string",
						BlobstoreDirectorPassword: "blobstore-director-some-random-string",
						BlobstoreAgentPassword:    "blobstore-agent-some-random-string",
						HMPassword:                "hm-some-random-string",
					},
				},
			))
		})

		It("does not generate an ssl keypair if it exists", func() {
			manifestProperties.SSLKeyPair = ssl.KeyPair{
				CA:          []byte(ca),
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}

			_, _, err := manifestBuilder.Build(manifestProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(sslKeyPairGenerator.GenerateCall.CallCount).To(Equal(0))
		})

		It("logs that the bosh-init manifest is being generated", func() {
			_, _, err := manifestBuilder.Build(manifestProperties)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Receives.Message).To(Equal("generating bosh-init manifest"))
		})

		It("stores the randomly generated passwords into manifest properties", func() {
			_, manifestProperties, err := manifestBuilder.Build(manifestProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(stringGenerator.GenerateCall.CallCount).To(Equal(14))
			Expect(manifestProperties.Credentials).To(Equal(manifests.InternalCredentials{
				MBusUsername:              "mbus-user-some-random-string",
				NatsUsername:              "nats-user-some-random-string",
				PostgresUsername:          "postgres-user-some-random-string",
				RegistryUsername:          "registry-user-some-random-string",
				BlobstoreDirectorUsername: "blobstore-director-user-some-random-string",
				BlobstoreAgentUsername:    "blobstore-agent-user-some-random-string",
				HMUsername:                "hm-user-some-random-string",
				MBusPassword:              "mbus-some-random-string",
				NatsPassword:              "nats-some-random-string",
				PostgresPassword:          "postgres-some-random-string",
				RegistryPassword:          "registry-some-random-string",
				BlobstoreDirectorPassword: "blobstore-director-some-random-string",
				BlobstoreAgentPassword:    "blobstore-agent-some-random-string",
				HMPassword:                "hm-some-random-string",
			}))
		})

		It("does not regenerate new random passwords if they already exist", func() {
			manifestProperties.Credentials = manifests.InternalCredentials{
				MBusUsername:              "mbus-user-some-persisted-string",
				NatsUsername:              "nats-user-some-persisted-string",
				PostgresUsername:          "postgres-user-some-persisted-string",
				RegistryUsername:          "registry-user-some-persisted-string",
				BlobstoreDirectorUsername: "blobstore-director-user-some-persisted-string",
				BlobstoreAgentUsername:    "blobstore-agent-user-some-persisted-string",
				HMUsername:                "hm-user-some-persisted-string",
				MBusPassword:              "mbus-some-persisted-string",
				NatsPassword:              "nats-some-persisted-string",
				PostgresPassword:          "postgres-some-persisted-string",
				RegistryPassword:          "registry-some-persisted-string",
				BlobstoreDirectorPassword: "blobstore-director-some-persisted-string",
				BlobstoreAgentPassword:    "blobstore-agent-some-persisted-string",
				HMPassword:                "hm-some-persisted-string",
			}

			_, manifestProperties, err := manifestBuilder.Build(manifestProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(stringGenerator.GenerateCall.CallCount).To(Equal(0))
			Expect(manifestProperties.Credentials).To(Equal(manifests.InternalCredentials{
				MBusUsername:              "mbus-user-some-persisted-string",
				NatsUsername:              "nats-user-some-persisted-string",
				PostgresUsername:          "postgres-user-some-persisted-string",
				RegistryUsername:          "registry-user-some-persisted-string",
				BlobstoreDirectorUsername: "blobstore-director-user-some-persisted-string",
				BlobstoreAgentUsername:    "blobstore-agent-user-some-persisted-string",
				HMUsername:                "hm-user-some-persisted-string",
				MBusPassword:              "mbus-some-persisted-string",
				NatsPassword:              "nats-some-persisted-string",
				PostgresPassword:          "postgres-some-persisted-string",
				RegistryPassword:          "registry-some-persisted-string",
				BlobstoreDirectorPassword: "blobstore-director-some-persisted-string",
				BlobstoreAgentPassword:    "blobstore-agent-some-persisted-string",
				HMPassword:                "hm-some-persisted-string",
			}))
		})

		Context("failure cases", func() {
			It("returns an error when the ssl key pair cannot be generated", func() {
				sslKeyPairGenerator.GenerateCall.Returns.Error = errors.New("failed to generate key pair")

				_, _, err := manifestBuilder.Build(manifestProperties)
				Expect(err).To(MatchError("failed to generate key pair"))
			})

			Context("failing cloud provider manifest builder", func() {
				BeforeEach(func() {
					fakeCloudProviderManifestBuilder := &fakes.CloudProviderManifestBuilder{}
					fakeCloudProviderManifestBuilder.BuildCall.Returns.Error = fmt.Errorf("something bad happened")
					manifestBuilder = manifests.NewManifestBuilder(logger, sslKeyPairGenerator, stringGenerator, fakeCloudProviderManifestBuilder, jobsManifestBuilder)
					manifestProperties = manifests.ManifestProperties{
						DirectorUsername: "bosh-username",
						DirectorPassword: "bosh-password",
						SubnetID:         "subnet-12345",
						AvailabilityZone: "some-az",
						CACommonName:     "BOSH Bootloader",
						ElasticIP:        "52.0.112.12",
						AccessKeyID:      "some-access-key-id",
						SecretAccessKey:  "some-secret-access-key",
						DefaultKeyName:   "some-key-name",
						Region:           "some-region",
						SecurityGroup:    "some-security-group",
					}
				})
				It("returns an error when it cannot build the cloud provider manifest", func() {
					_, _, err := manifestBuilder.Build(manifestProperties)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("failing jobs manifest builder", func() {
				BeforeEach(func() {
					fakeJobsManifestBuilder := &fakes.JobsManifestBuilder{}
					fakeJobsManifestBuilder.BuildCall.Returns.Error = fmt.Errorf("something bad happened")
					manifestBuilder = manifests.NewManifestBuilder(logger, sslKeyPairGenerator, stringGenerator, cloudProviderManifestBuilder, fakeJobsManifestBuilder)
					manifestProperties = manifests.ManifestProperties{
						DirectorUsername: "bosh-username",
						DirectorPassword: "bosh-password",
						SubnetID:         "subnet-12345",
						AvailabilityZone: "some-az",
						CACommonName:     "BOSH Bootloader",
						ElasticIP:        "52.0.112.12",
						AccessKeyID:      "some-access-key-id",
						SecretAccessKey:  "some-secret-access-key",
						DefaultKeyName:   "some-key-name",
						Region:           "some-region",
						SecurityGroup:    "some-security-group",
					}
				})
				It("returns an error when it cannot build the job manifest", func() {
					_, _, err := manifestBuilder.Build(manifestProperties)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("template marshaling", func() {
		It("can be marshaled to YML", func() {
			sslKeyPairGenerator.GenerateCall.Returns.KeyPair = ssl.KeyPair{
				CA:          []byte(ca),
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}

			manifest, _, err := manifestBuilder.Build(manifestProperties)
			Expect(err).NotTo(HaveOccurred())

			buf, err := ioutil.ReadFile("fixtures/manifest.yml")
			Expect(err).NotTo(HaveOccurred())

			output, err := yaml.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchYAML(string(buf)))
		})
	})
})
