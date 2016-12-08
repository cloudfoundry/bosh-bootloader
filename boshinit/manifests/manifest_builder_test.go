package manifests_test

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/ssl"

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
		awsManifestProperties        manifests.ManifestProperties
		gcpManifestProperties        manifests.ManifestProperties
		cloudProviderManifestBuilder manifests.CloudProviderManifestBuilder
		jobsManifestBuilder          manifests.JobsManifestBuilder
		input                        manifests.ManifestBuilderInput
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		sslKeyPairGenerator = &fakes.SSLKeyPairGenerator{}
		stringGenerator = &fakes.StringGenerator{}
		cloudProviderManifestBuilder = manifests.NewCloudProviderManifestBuilder(stringGenerator)
		jobsManifestBuilder = manifests.NewJobsManifestBuilder(stringGenerator)
		input = manifests.ManifestBuilderInput{
			AWSBOSHURL:      "some-aws-bosh-url",
			AWSBOSHSHA1:     "some-aws-bosh-sha1",
			GCPBOSHURL:      "some-google-bosh-url",
			GCPBOSHSHA1:     "some-google-bosh-sha1",
			BOSHAWSCPIURL:   "some-bosh-aws-cpi-url",
			BOSHAWSCPISHA1:  "some-bosh-aws-cpi-sha1",
			BOSHGCPCPIURL:   "some-bosh-google-cpi-url",
			BOSHGCPCPISHA1:  "some-bosh-google-cpi-sha1",
			AWSStemcellURL:  "some-aws-stemcell-url",
			AWSStemcellSHA1: "some-aws-stemcell-sha1",
			GCPStemcellURL:  "some-google-stemcell-url",
			GCPStemcellSHA1: "some-google-stemcell-sha1",
		}

		manifestBuilder = manifests.NewManifestBuilder(input, logger, sslKeyPairGenerator, stringGenerator, cloudProviderManifestBuilder, jobsManifestBuilder)
		awsManifestProperties = manifests.ManifestProperties{
			DirectorName:     "bosh-name",
			DirectorUsername: "bosh-username",
			DirectorPassword: "bosh-password",
			CACommonName:     "BOSH Bootloader",
			ExternalIP:       "52.0.112.12",
			AWS: manifests.ManifestPropertiesAWS{
				SubnetID:         "subnet-12345",
				AvailabilityZone: "some-az",
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				DefaultKeyName:   "some-key-name",
				Region:           "some-region",
				SecurityGroup:    "some-security-group",
			},
		}

		gcpManifestProperties = manifests.ManifestProperties{
			DirectorName:     "bosh-name",
			DirectorUsername: "bosh-username",
			DirectorPassword: "bosh-password",
			ExternalIP:       "52.0.112.12",
			CACommonName:     "BOSH Bootloader",
			GCP: manifests.ManifestPropertiesGCP{
				Zone:           "some-zone",
				NetworkName:    "some-network-name",
				SubnetworkName: "some-subnet-name",
				BOSHTag:        "some-bosh-tag",
				InternalTag:    "some-internal-tag",
				Project:        "some-project",
				JsonKey: `{
  "type": "service_account",
  "project_id": "some-project",
  "private_key_id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "private_key": "-----BEGIN PRIVATE KEY-----\nxxxx=\n-----END PRIVATE KEY-----\n",
  "client_email": "test-account@some-project.iam.gserviceaccount.com",
  "client_id": "xxxxxxxxxxxxxxxxxxxxx",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test-account%40some-project.iam.gserviceaccount.com"
}
`,
			},
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

			manifest, awsManifestProperties, err := manifestBuilder.Build("aws", awsManifestProperties)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest.Name).To(Equal("bosh"))
			Expect(manifest.Releases[0].Name).To(Equal("bosh"))
			Expect(manifest.ResourcePools[0].CloudProperties.AvailabilityZone).To(Equal("some-az"))
			Expect(manifest.DiskPools[0].Name).To(Equal("disks"))
			Expect(manifest.Networks[0].Subnets[0].CloudProperties.Subnet).To(Equal("subnet-12345"))
			Expect(manifest.Jobs[0].Networks[1].StaticIPs[0]).To(Equal("52.0.112.12"))
			Expect(manifest.Jobs[0].Properties.Director.Name).To(Equal("bosh-name"))
			Expect(manifest.Jobs[0].Properties.Director.SSL).To(Equal(manifests.SSLProperties{
				Cert: certificate,
				Key:  privateKey,
			}))
			Expect(manifest.CloudProvider.SSHTunnel.Host).To(Equal("52.0.112.12"))
			Expect(manifest.CloudProvider.MBus).To(Equal("https://mbus-user-some-random-string:mbus-some-random-string@52.0.112.12:6868"))

			Expect(sslKeyPairGenerator.GenerateCall.Receives.CACommonName).To(Equal("BOSH Bootloader"))
			Expect(sslKeyPairGenerator.GenerateCall.Receives.CertCommonName).To(Equal("52.0.112.12"))
			Expect(sslKeyPairGenerator.GenerateCall.CallCount).To(Equal(1))

			Expect(awsManifestProperties).To(Equal(
				manifests.ManifestProperties{
					DirectorName:     "bosh-name",
					DirectorUsername: "bosh-username",
					DirectorPassword: "bosh-password",
					CACommonName:     "BOSH Bootloader",
					ExternalIP:       "52.0.112.12",
					AWS: manifests.ManifestPropertiesAWS{
						SubnetID:         "subnet-12345",
						AvailabilityZone: "some-az",
						AccessKeyID:      "some-access-key-id",
						SecretAccessKey:  "some-secret-access-key",
						DefaultKeyName:   "some-key-name",
						Region:           "some-region",
						SecurityGroup:    "some-security-group",
					},
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

		It("builds the bosh-init manifest and updates the manifest properties for aws", func() {
			sslKeyPairGenerator.GenerateCall.Returns.KeyPair = ssl.KeyPair{
				CA:          []byte(ca),
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}

			manifest, _, err := manifestBuilder.Build("aws", awsManifestProperties)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest.Releases[0]).To(Equal(manifests.Release{
				Name: "bosh",
				URL:  "some-aws-bosh-url",
				SHA1: "some-aws-bosh-sha1"},
			))

			Expect(manifest.Releases[1]).To(Equal(manifests.Release{
				Name: "bosh-aws-cpi",
				URL:  "some-bosh-aws-cpi-url",
				SHA1: "some-bosh-aws-cpi-sha1",
			}))

			Expect(manifest.ResourcePools[0].Stemcell).To(Equal(manifests.Stemcell{
				URL:  "some-aws-stemcell-url",
				SHA1: "some-aws-stemcell-sha1",
			}))

			expectedAWSProperties := manifests.AWSProperties{
				AccessKeyId:           "some-access-key-id",
				SecretAccessKey:       "some-secret-access-key",
				DefaultKeyName:        "some-key-name",
				DefaultSecurityGroups: []string{"some-security-group"},
				Region:                "some-region",
			}
			Expect(manifest.Jobs[0].Properties.AWS).To(Equal(expectedAWSProperties))
			Expect(manifest.CloudProvider.Properties.AWS).To(Equal(expectedAWSProperties))
		})

		It("builds the bosh-init manifest and updates the manifest properties for gcp", func() {
			sslKeyPairGenerator.GenerateCall.Returns.KeyPair = ssl.KeyPair{
				CA:          []byte(ca),
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}

			manifest, _, err := manifestBuilder.Build("gcp", gcpManifestProperties)
			Expect(err).NotTo(HaveOccurred())

			Expect(manifest.Releases[0]).To(Equal(manifests.Release{
				Name: "bosh",
				URL:  "some-google-bosh-url",
				SHA1: "some-google-bosh-sha1"},
			))

			Expect(manifest.Releases[1]).To(Equal(manifests.Release{
				Name: "bosh-google-cpi",
				URL:  "some-bosh-google-cpi-url",
				SHA1: "some-bosh-google-cpi-sha1",
			}))

			Expect(manifest.ResourcePools[0].Stemcell).To(Equal(manifests.Stemcell{
				URL:  "some-google-stemcell-url",
				SHA1: "some-google-stemcell-sha1",
			}))

			Expect(manifest.Jobs[0].Properties.Google).To(Equal(manifests.GoogleProperties{
				Project: "some-project",
				JsonKey: `{
  "type": "service_account",
  "project_id": "some-project",
  "private_key_id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "private_key": "-----BEGIN PRIVATE KEY-----\nxxxx=\n-----END PRIVATE KEY-----\n",
  "client_email": "test-account@some-project.iam.gserviceaccount.com",
  "client_id": "xxxxxxxxxxxxxxxxxxxxx",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test-account%40some-project.iam.gserviceaccount.com"
}
`,
			}))
		})

		It("does not generate an ssl keypair if it exists", func() {
			awsManifestProperties.SSLKeyPair = ssl.KeyPair{
				CA:          []byte(ca),
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}

			_, _, err := manifestBuilder.Build("aws", awsManifestProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(sslKeyPairGenerator.GenerateCall.CallCount).To(Equal(0))
		})

		It("logs that the bosh-init manifest is being generated", func() {
			_, _, err := manifestBuilder.Build("aws", awsManifestProperties)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Receives.Message).To(Equal("generating bosh-init manifest"))
		})

		It("stores the randomly generated passwords into manifest properties", func() {
			_, awsManifestProperties, err := manifestBuilder.Build("aws", awsManifestProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(stringGenerator.GenerateCall.CallCount).To(Equal(14))
			Expect(awsManifestProperties.Credentials).To(Equal(manifests.InternalCredentials{
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
			awsManifestProperties.Credentials = manifests.InternalCredentials{
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

			_, awsManifestProperties, err := manifestBuilder.Build("aws", awsManifestProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(stringGenerator.GenerateCall.CallCount).To(Equal(0))
			Expect(awsManifestProperties.Credentials).To(Equal(manifests.InternalCredentials{
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

				_, _, err := manifestBuilder.Build("aws", awsManifestProperties)
				Expect(err).To(MatchError("failed to generate key pair"))
			})

			Context("failing cloud provider manifest builder", func() {
				BeforeEach(func() {
					fakeCloudProviderManifestBuilder := &fakes.CloudProviderManifestBuilder{}
					fakeCloudProviderManifestBuilder.BuildCall.Returns.Error = fmt.Errorf("something bad happened")
					manifestBuilder = manifests.NewManifestBuilder(input, logger, sslKeyPairGenerator, stringGenerator, fakeCloudProviderManifestBuilder, jobsManifestBuilder)
					awsManifestProperties = manifests.ManifestProperties{
						DirectorUsername: "bosh-username",
						DirectorPassword: "bosh-password",
						CACommonName:     "BOSH Bootloader",
						ExternalIP:       "52.0.112.12",
						AWS: manifests.ManifestPropertiesAWS{
							SubnetID:         "subnet-12345",
							AvailabilityZone: "some-az",
							AccessKeyID:      "some-access-key-id",
							SecretAccessKey:  "some-secret-access-key",
							DefaultKeyName:   "some-key-name",
							Region:           "some-region",
							SecurityGroup:    "some-security-group",
						},
					}
				})
				It("returns an error when it cannot build the cloud provider manifest", func() {
					_, _, err := manifestBuilder.Build("aws", awsManifestProperties)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("failing jobs manifest builder", func() {
				BeforeEach(func() {
					fakeJobsManifestBuilder := &fakes.JobsManifestBuilder{}
					fakeJobsManifestBuilder.BuildCall.Returns.Error = fmt.Errorf("something bad happened")
					manifestBuilder = manifests.NewManifestBuilder(input, logger, sslKeyPairGenerator, stringGenerator, cloudProviderManifestBuilder, fakeJobsManifestBuilder)
					awsManifestProperties = manifests.ManifestProperties{
						DirectorUsername: "bosh-username",
						DirectorPassword: "bosh-password",
						CACommonName:     "BOSH Bootloader",
						ExternalIP:       "52.0.112.12",
						AWS: manifests.ManifestPropertiesAWS{
							SubnetID:         "subnet-12345",
							AvailabilityZone: "some-az",
							AccessKeyID:      "some-access-key-id",
							SecretAccessKey:  "some-secret-access-key",
							DefaultKeyName:   "some-key-name",
							Region:           "some-region",
							SecurityGroup:    "some-security-group",
						},
					}
				})
				It("returns an error when it cannot build the job manifest", func() {
					_, _, err := manifestBuilder.Build("aws", awsManifestProperties)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("template marshaling", func() {
		BeforeEach(func() {
			sslKeyPairGenerator.GenerateCall.Returns.KeyPair = ssl.KeyPair{
				CA:          []byte(ca),
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}
		})

		It("can correctly marshal an aws manifest", func() {
			manifest, _, err := manifestBuilder.Build("aws", awsManifestProperties)
			Expect(err).NotTo(HaveOccurred())

			buf, err := ioutil.ReadFile("fixtures/aws_manifest.yml")
			Expect(err).NotTo(HaveOccurred())

			output, err := yaml.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchYAML(string(buf)))
		})

		It("can correctly marshal an gcp manifest", func() {
			manifest, _, err := manifestBuilder.Build("gcp", gcpManifestProperties)
			Expect(err).NotTo(HaveOccurred())

			buf, err := ioutil.ReadFile("fixtures/gcp_manifest.yml")
			Expect(err).NotTo(HaveOccurred())

			output, err := yaml.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchYAML(string(buf)))
		})
	})

})
