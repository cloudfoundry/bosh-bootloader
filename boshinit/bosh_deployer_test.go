package boshinit_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoshDeployer", func() {
	var (
		manifestBuilder             *fakes.BOSHInitManifestBuilder
		boshInitRunner              *fakes.BOSHInitRunner
		boshDeployer                boshinit.BOSHDeployer
		logger                      *fakes.Logger
		infrastructureConfiguration boshinit.InfrastructureConfiguration
		sslKeyPair                  ssl.KeyPair
		ec2KeyPair                  ec2.KeyPair
		credentials                 map[string]string
	)

	BeforeEach(func() {
		manifestBuilder = &fakes.BOSHInitManifestBuilder{}
		boshInitRunner = &fakes.BOSHInitRunner{}
		logger = &fakes.Logger{}
		boshDeployer = boshinit.NewBOSHDeployer(manifestBuilder, boshInitRunner, logger)

		infrastructureConfiguration = boshinit.InfrastructureConfiguration{
			SubnetID:         "subnet-12345",
			AvailabilityZone: "some-az",
			ElasticIP:        "some-elastic-ip",
			AccessKeyID:      "some-access-key-id",
			SecretAccessKey:  "some-secret-access-key",
			SecurityGroup:    "some-security-group",
			AWSRegion:        "some-aws-region",
		}

		sslKeyPair = ssl.KeyPair{
			Certificate: []byte("some-certificate"),
			PrivateKey:  []byte("some-private-key"),
		}

		ec2KeyPair = ec2.KeyPair{
			Name:       "some-keypair-name",
			PrivateKey: "some-private-key",
			PublicKey:  "some-public-key",
		}

		credentials = map[string]string{
			"mbusUsername":              "some-mbus-username",
			"natsUsername":              "some-nats-username",
			"postgresUsername":          "some-postgres-username",
			"registryUsername":          "some-registry-username",
			"blobstoreDirectorUsername": "some-blobstore-director-username",
			"blobstoreAgentUsername":    "some-blobstore-agent-username",
			"hmUsername":                "some-hm-username",
			"mbusPassword":              "some-mbus-password",
			"natsPassword":              "some-nats-password",
			"redisPassword":             "some-redis-password",
			"postgresPassword":          "some-postgres-password",
			"registryPassword":          "some-registry-password",
			"blobstoreDirectorPassword": "some-blobstore-director-password",
			"blobstoreAgentPassword":    "some-blobstore-agent-password",
			"hmPassword":                "some-hm-password",
		}

		manifestBuilder.BuildCall.Returns.Properties = manifests.ManifestProperties{
			DirectorUsername: "admin",
			DirectorPassword: "admin",
			ElasticIP:        "some-elastic-ip",
			SSLKeyPair: ssl.KeyPair{
				Certificate: []byte("updated-certificate"),
				PrivateKey:  []byte("updated-private-key"),
			},
			Credentials: manifests.InternalCredentials{
				MBusUsername:              "some-mbus-username",
				NatsUsername:              "some-nats-username",
				PostgresUsername:          "some-postgres-username",
				RegistryUsername:          "some-registry-username",
				BlobstoreDirectorUsername: "some-blobstore-director-username",
				BlobstoreAgentUsername:    "some-blobstore-agent-username",
				HMUsername:                "some-hm-username",
				MBusPassword:              "some-mbus-password",
				NatsPassword:              "some-nats-password",
				RedisPassword:             "some-redis-password",
				PostgresPassword:          "some-postgres-password",
				RegistryPassword:          "some-registry-password",
				BlobstoreDirectorPassword: "some-blobstore-director-password",
				BlobstoreAgentPassword:    "some-blobstore-agent-password",
				HMPassword:                "some-hm-password",
			},
		}
		manifestBuilder.BuildCall.Returns.Manifest = manifests.Manifest{
			Name: "bosh",
		}

		boshInitRunner.DeployCall.Returns.State = boshinit.State{
			"key": "value",
		}
	})

	Describe("Deploy", func() {
		It("deploys bosh and returns a bosh output", func() {
			boshOutput, err := boshDeployer.Deploy(boshinit.BOSHDeployInput{
				DirectorUsername: "some-director",
				DirectorPassword: "some-password",
				State: boshinit.State{
					"key": "value",
				},
				InfrastructureConfiguration: infrastructureConfiguration,
				SSLKeyPair:                  sslKeyPair,
				EC2KeyPair:                  ec2KeyPair,
				Credentials:                 credentials,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(manifestBuilder.BuildCall.Receives.Properties).To(Equal(manifests.ManifestProperties{
				DirectorUsername: "some-director",
				DirectorPassword: "some-password",
				SubnetID:         "subnet-12345",
				AvailabilityZone: "some-az",
				ElasticIP:        "some-elastic-ip",
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				DefaultKeyName:   "some-keypair-name",
				Region:           "some-aws-region",
				SecurityGroup:    "some-security-group",
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("some-certificate"),
					PrivateKey:  []byte("some-private-key"),
				},
				Credentials: manifests.InternalCredentials{
					MBusUsername:              "some-mbus-username",
					NatsUsername:              "some-nats-username",
					PostgresUsername:          "some-postgres-username",
					RegistryUsername:          "some-registry-username",
					BlobstoreDirectorUsername: "some-blobstore-director-username",
					BlobstoreAgentUsername:    "some-blobstore-agent-username",
					HMUsername:                "some-hm-username",
					MBusPassword:              "some-mbus-password",
					NatsPassword:              "some-nats-password",
					RedisPassword:             "some-redis-password",
					PostgresPassword:          "some-postgres-password",
					RegistryPassword:          "some-registry-password",
					BlobstoreDirectorPassword: "some-blobstore-director-password",
					BlobstoreAgentPassword:    "some-blobstore-agent-password",
					HMPassword:                "some-hm-password",
				},
			}))
			Expect(boshOutput.DirectorSSLKeyPair).To(Equal(ssl.KeyPair{
				Certificate: []byte("updated-certificate"),
				PrivateKey:  []byte("updated-private-key"),
			}))
			Expect(boshOutput.BOSHInitState).To(Equal(boshinit.State{
				"key": "value",
			}))
			Expect(boshOutput.Credentials).To(Equal(map[string]string{
				"mbusUsername":              "some-mbus-username",
				"natsUsername":              "some-nats-username",
				"postgresUsername":          "some-postgres-username",
				"registryUsername":          "some-registry-username",
				"blobstoreDirectorUsername": "some-blobstore-director-username",
				"blobstoreAgentUsername":    "some-blobstore-agent-username",
				"hmUsername":                "some-hm-username",
				"mbusPassword":              "some-mbus-password",
				"natsPassword":              "some-nats-password",
				"redisPassword":             "some-redis-password",
				"postgresPassword":          "some-postgres-password",
				"registryPassword":          "some-registry-password",
				"blobstoreDirectorPassword": "some-blobstore-director-password",
				"blobstoreAgentPassword":    "some-blobstore-agent-password",
				"hmPassword":                "some-hm-password",
			}))

			Expect(boshInitRunner.DeployCall.Receives.Manifest).To(ContainSubstring("name: bosh"))
			Expect(boshInitRunner.DeployCall.Receives.PrivateKey).To(ContainSubstring("some-private-key"))
			Expect(boshInitRunner.DeployCall.Receives.State).To(Equal(boshinit.State{
				"key": "value",
			}))
		})

		It("prints out the bosh director information", func() {
			var lines []string
			logger.PrintlnCall.Stub = func(line string) {
				lines = append(lines, line)
			}

			_, err := boshDeployer.Deploy(boshinit.BOSHDeployInput{
				InfrastructureConfiguration: infrastructureConfiguration,
				SSLKeyPair:                  sslKeyPair,
				EC2KeyPair:                  ec2KeyPair,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(lines).To(ContainElement("Director Address:  https://some-elastic-ip:25555"))
			Expect(lines).To(ContainElement("Director Username: admin"))
			Expect(lines).To(ContainElement("Director Password: admin"))
		})

		Context("failure cases", func() {
			Context("when the manifest cannot be built", func() {
				It("returns an error", func() {
					manifestBuilder.BuildCall.Returns.Error = errors.New("failed to build manifest")

					_, err := boshDeployer.Deploy(boshinit.BOSHDeployInput{})
					Expect(err).To(MatchError("failed to build manifest"))
				})
			})

			Context("when the runner fails to deploy", func() {
				It("returns an error", func() {
					boshInitRunner.DeployCall.Returns.Error = errors.New("failed to deploy")

					_, err := boshDeployer.Deploy(boshinit.BOSHDeployInput{})
					Expect(err).To(MatchError("failed to deploy"))
				})
			})
		})
	})
})
