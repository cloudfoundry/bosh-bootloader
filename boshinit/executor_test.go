package boshinit_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/ssl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	var (
		manifestBuilder             *fakes.BOSHInitManifestBuilder
		deployCommandRunner         *fakes.BOSHInitCommandRunner
		deleteCommandRunner         *fakes.BOSHInitCommandRunner
		executor                    boshinit.Executor
		logger                      *fakes.Logger
		infrastructureConfiguration boshinit.InfrastructureConfiguration
		sslKeyPair                  ssl.KeyPair
		ec2KeyPair                  ec2.KeyPair
		credentials                 map[string]string
	)

	BeforeEach(func() {
		manifestBuilder = &fakes.BOSHInitManifestBuilder{}
		deployCommandRunner = &fakes.BOSHInitCommandRunner{}
		deleteCommandRunner = &fakes.BOSHInitCommandRunner{}
		logger = &fakes.Logger{}
		executor = boshinit.NewExecutor(manifestBuilder, deployCommandRunner, deleteCommandRunner, logger)

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
			"postgresPassword":          "some-postgres-password",
			"registryPassword":          "some-registry-password",
			"blobstoreDirectorPassword": "some-blobstore-director-password",
			"blobstoreAgentPassword":    "some-blobstore-agent-password",
			"hmPassword":                "some-hm-password",
		}

		manifestBuilder.BuildCall.Returns.Properties = manifests.ManifestProperties{
			DirectorUsername: "admin",
			DirectorPassword: "admin",
			CACommonName:     "BOSH Bootloader",
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

		deployCommandRunner.ExecuteCall.Returns.State = boshinit.State{
			"key": "value",
		}
	})

	Describe("Delete", func() {
		It("deletes the bosh director given the state", func() {
			err := executor.Delete("bosh-init-manifest", boshinit.State{"key": "value"}, "ec2-private-key")
			Expect(err).NotTo(HaveOccurred())

			Expect(deleteCommandRunner.ExecuteCall.Receives.Manifest).To(Equal([]byte("bosh-init-manifest")))
			Expect(deleteCommandRunner.ExecuteCall.Receives.PrivateKey).To(Equal("ec2-private-key"))
			Expect(deleteCommandRunner.ExecuteCall.Receives.State).To(Equal(boshinit.State{"key": "value"}))
		})

		It("prints out that the director is being destroyed", func() {
			err := executor.Delete("", boshinit.State{}, "")
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Receives.Message).To(Equal("destroying bosh director"))
		})

		Context("failure cases", func() {
			Context("when the runner fails to delete", func() {
				It("returns an error", func() {
					deleteCommandRunner.ExecuteCall.Returns.Error = errors.New("failed to delete")

					err := executor.Delete("", boshinit.State{}, "")
					Expect(err).To(MatchError("failed to delete"))
				})
			})
		})
	})

	Describe("Deploy", func() {
		It("deploys bosh and returns a bosh output", func() {
			deployOutput, err := executor.Deploy(boshinit.DeployInput{
				DirectorName:     "some-director-name",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
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
				DirectorName:     "some-director-name",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				SubnetID:         "subnet-12345",
				AvailabilityZone: "some-az",
				CACommonName:     "BOSH Bootloader",
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
					PostgresPassword:          "some-postgres-password",
					RegistryPassword:          "some-registry-password",
					BlobstoreDirectorPassword: "some-blobstore-director-password",
					BlobstoreAgentPassword:    "some-blobstore-agent-password",
					HMPassword:                "some-hm-password",
				},
			}))

			Expect(deployOutput.DirectorSSLKeyPair).To(Equal(ssl.KeyPair{
				Certificate: []byte("updated-certificate"),
				PrivateKey:  []byte("updated-private-key"),
			}))

			Expect(deployOutput.BOSHInitState).To(Equal(boshinit.State{
				"key": "value",
			}))

			Expect(deployOutput.Credentials).To(Equal(map[string]string{
				"mbusUsername":              "some-mbus-username",
				"natsUsername":              "some-nats-username",
				"postgresUsername":          "some-postgres-username",
				"registryUsername":          "some-registry-username",
				"blobstoreDirectorUsername": "some-blobstore-director-username",
				"blobstoreAgentUsername":    "some-blobstore-agent-username",
				"hmUsername":                "some-hm-username",
				"mbusPassword":              "some-mbus-password",
				"natsPassword":              "some-nats-password",
				"postgresPassword":          "some-postgres-password",
				"registryPassword":          "some-registry-password",
				"blobstoreDirectorPassword": "some-blobstore-director-password",
				"blobstoreAgentPassword":    "some-blobstore-agent-password",
				"hmPassword":                "some-hm-password",
			}))

			Expect(deployOutput.BOSHInitManifest).To(ContainSubstring("name: bosh"))

			Expect(deployCommandRunner.ExecuteCall.Receives.Manifest).To(ContainSubstring("name: bosh"))
			Expect(deployCommandRunner.ExecuteCall.Receives.PrivateKey).To(ContainSubstring("some-private-key"))
			Expect(deployCommandRunner.ExecuteCall.Receives.State).To(Equal(boshinit.State{
				"key": "value",
			}))
		})

		It("prints out that the director is being deployed", func() {
			_, err := executor.Deploy(boshinit.DeployInput{
				InfrastructureConfiguration: infrastructureConfiguration,
				SSLKeyPair:                  sslKeyPair,
				EC2KeyPair:                  ec2KeyPair,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Receives.Message).To(Equal("deploying bosh director"))
		})

		Context("failure cases", func() {
			Context("when the manifest cannot be built", func() {
				It("returns an error", func() {
					manifestBuilder.BuildCall.Returns.Error = errors.New("failed to build manifest")

					_, err := executor.Deploy(boshinit.DeployInput{})
					Expect(err).To(MatchError("failed to build manifest"))
				})
			})

			Context("when the runner fails to deploy", func() {
				It("returns an error", func() {
					deployCommandRunner.ExecuteCall.Returns.Error = errors.New("failed to deploy")

					_, err := executor.Deploy(boshinit.DeployInput{})
					Expect(err).To(MatchError("failed to deploy"))
				})
			})
		})
	})
})
