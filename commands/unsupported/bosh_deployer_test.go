package unsupported_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoshDeployer", func() {
	var (
		manifestBuilder *fakes.BOSHInitManifestBuilder
		boshInitRunner  *fakes.BOSHInitRunner
		boshDeployer    unsupported.BOSHDeployer
		logger          *fakes.Logger
		stack           cloudformation.Stack
		sslKeyPair      ssl.KeyPair
		ec2KeyPair      ec2.KeyPair
		credentials     boshinit.InternalCredentials
	)

	BeforeEach(func() {
		manifestBuilder = &fakes.BOSHInitManifestBuilder{}
		boshInitRunner = &fakes.BOSHInitRunner{}
		logger = &fakes.Logger{}
		boshDeployer = unsupported.NewBOSHDeployer(manifestBuilder, boshInitRunner, logger)

		stack = cloudformation.Stack{
			Outputs: map[string]string{
				"BOSHSubnet":              "subnet-12345",
				"BOSHSubnetAZ":            "some-az",
				"BOSHEIP":                 "some-elastic-ip",
				"BOSHUserAccessKey":       "some-access-key-id",
				"BOSHUserSecretAccessKey": "some-secret-access-key",
				"BOSHSecurityGroup":       "some-security-group",
			},
		}
		sslKeyPair = ssl.KeyPair{
			Certificate: []byte("some-certificate"),
			PrivateKey:  []byte("some-private-key"),
		}
		ec2KeyPair = ec2.KeyPair{
			Name:       "some-keypair-name",
			PrivateKey: []byte("some-private-key"),
			PublicKey:  []byte("some-public-key"),
		}
		credentials = boshinit.InternalCredentials{
			MBusPassword:              "some-mbus-password",
			NatsPassword:              "some-nats-password",
			RedisPassword:             "some-redis-password",
			PostgresPassword:          "some-postgres-password",
			RegistryPassword:          "some-registry-password",
			BlobstoreDirectorPassword: "some-blobstore-director-password",
			BlobstoreAgentPassword:    "some-blobstore-agent-password",
			HMPassword:                "some-hm-password",
		}

		manifestBuilder.BuildCall.Returns.Properties = boshinit.ManifestProperties{
			DirectorUsername: "admin",
			DirectorPassword: "admin",
			ElasticIP:        "some-elastic-ip",
			SSLKeyPair: ssl.KeyPair{
				Certificate: []byte("updated-certificate"),
				PrivateKey:  []byte("updated-private-key"),
			},
			Credentials: credentials,
		}
		manifestBuilder.BuildCall.Returns.Manifest = boshinit.Manifest{
			Name: "bosh",
		}

		boshInitRunner.DeployCall.Returns.State = boshinit.State{
			"key": "value",
		}
	})

	Describe("Deploy", func() {
		It("deploys bosh and returns a bosh output", func() {
			boshOutput, err := boshDeployer.Deploy(unsupported.BOSHDeployInput{
				State: boshinit.State{
					"key": "value",
				},
				Stack:       stack,
				AWSRegion:   "some-aws-region",
				SSLKeyPair:  sslKeyPair,
				EC2KeyPair:  ec2KeyPair,
				Credentials: credentials,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(manifestBuilder.BuildCall.Receives.Properties).To(Equal(boshinit.ManifestProperties{
				DirectorUsername: "admin",
				DirectorPassword: "admin",
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
				Credentials: boshinit.InternalCredentials{
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
			Expect(boshOutput.Credentials).To(Equal(boshinit.InternalCredentials{
				MBusPassword:              "some-mbus-password",
				NatsPassword:              "some-nats-password",
				RedisPassword:             "some-redis-password",
				PostgresPassword:          "some-postgres-password",
				RegistryPassword:          "some-registry-password",
				BlobstoreDirectorPassword: "some-blobstore-director-password",
				BlobstoreAgentPassword:    "some-blobstore-agent-password",
				HMPassword:                "some-hm-password",
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

			_, err := boshDeployer.Deploy(unsupported.BOSHDeployInput{
				Stack:      stack,
				AWSRegion:  "some-aws-region",
				SSLKeyPair: sslKeyPair,
				EC2KeyPair: ec2KeyPair,
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

					_, err := boshDeployer.Deploy(unsupported.BOSHDeployInput{})
					Expect(err).To(MatchError("failed to build manifest"))
				})
			})

			Context("when the runner fails to deploy", func() {
				It("returns an error", func() {
					boshInitRunner.DeployCall.Returns.Error = errors.New("failed to deploy")

					_, err := boshDeployer.Deploy(unsupported.BOSHDeployInput{})
					Expect(err).To(MatchError("failed to deploy"))
				})
			})
		})
	})
})
