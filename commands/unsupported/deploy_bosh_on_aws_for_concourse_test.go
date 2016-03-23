package unsupported_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeployBOSHOnAWSForConcourse", func() {
	Describe("Execute", func() {
		var (
			command                   unsupported.DeployBOSHOnAWSForConcourse
			boshDeployer              *fakes.BOSHDeployer
			infrastructureCreator     *fakes.InfrastructureCreator
			keyPairSynchronizer       *fakes.KeyPairSynchronizer
			cloudFormationClient      *fakes.CloudFormationClient
			ec2Client                 *fakes.EC2Client
			clientProvider            *fakes.ClientProvider
			stringGenerator           *fakes.StringGenerator
			cloudConfigurator         *fakes.BoshCloudConfigurator
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			incomingState             storage.State
			globalFlags               commands.GlobalFlags
		)

		BeforeEach(func() {
			keyPairSynchronizer = &fakes.KeyPairSynchronizer{}
			keyPairSynchronizer.SyncCall.Returns.KeyPair = unsupported.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}

			infrastructureCreator = &fakes.InfrastructureCreator{}
			infrastructureCreator.CreateCall.Returns.Stack = cloudformation.Stack{
				Name: "concourse",
			}

			boshDeployer = &fakes.BOSHDeployer{}
			boshDeployer.DeployCall.Returns.Output = unsupported.BOSHDeployOutput{
				DirectorSSLKeyPair: ssl.KeyPair{
					Certificate: []byte("updated-certificate"),
					PrivateKey:  []byte("updated-private-key"),
				},
				BOSHInitState: boshinit.State{
					"updated-key": "updated-value",
				},
			}

			cloudFormationClient = &fakes.CloudFormationClient{}
			ec2Client = &fakes.EC2Client{}

			clientProvider = &fakes.ClientProvider{}
			clientProvider.CloudFormationClientCall.Returns.Client = cloudFormationClient
			clientProvider.EC2ClientCall.Returns.Client = ec2Client

			stringGenerator = &fakes.StringGenerator{}
			stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				return fmt.Sprintf("%s%s", prefix, "some-random-string"), nil
			}

			cloudConfigurator = &fakes.BoshCloudConfigurator{}

			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}

			globalFlags = commands.GlobalFlags{
				EndpointOverride: "some-endpoint",
			}

			incomingState = storage.State{
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				KeyPair: &storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: &storage.BOSH{
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					State: map[string]interface{}{
						"key": "value",
					},
				},
			}

			command = unsupported.NewDeployBOSHOnAWSForConcourse(
				infrastructureCreator, keyPairSynchronizer, clientProvider, boshDeployer,
				stringGenerator, cloudConfigurator, availabilityZoneRetriever)
		})

		It("syncs the keypair", func() {
			state, err := command.Execute(globalFlags, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(clientProvider.EC2ClientCall.Receives.Config).To(Equal(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint",
			}))
			Expect(keyPairSynchronizer.SyncCall.Receives.EC2Client).To(Equal(ec2Client))
			Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(unsupported.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}))

			Expect(state.KeyPair).To(Equal(&storage.KeyPair{
				Name:       "some-keypair-name",
				PublicKey:  "some-public-key",
				PrivateKey: "some-private-key",
			}))
		})

		It("creates/updates the stack with the given name", func() {
			_, err := command.Execute(globalFlags, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(clientProvider.CloudFormationClientCall.Receives.Config).To(Equal(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint",
			}))
			Expect(infrastructureCreator.CreateCall.Receives.KeyPairName).To(Equal("some-keypair-name"))
			Expect(infrastructureCreator.CreateCall.Returns.Error).To(BeNil())
		})

		It("deploys bosh", func() {
			_, err := command.Execute(globalFlags, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshDeployer.DeployCall.Receives.Input).To(Equal(unsupported.BOSHDeployInput{
				DirectorUsername: "user-some-random-string",
				DirectorPassword: "p-some-random-string",
				State: boshinit.State{
					"key": "value",
				},
				Stack: cloudformation.Stack{
					Name: "concourse",
				},
				AWSRegion: "some-aws-region",
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("some-certificate"),
					PrivateKey:  []byte("some-private-key"),
				},
				EC2KeyPair: ec2.KeyPair{
					Name:       "some-keypair-name",
					PublicKey:  []byte("some-public-key"),
					PrivateKey: []byte("some-private-key"),
				},
			}))
		})

		Context("when there is no keypair", func() {
			BeforeEach(func() {
				incomingState.KeyPair = nil
			})

			It("syncs with an empty keypair", func() {
				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairSynchronizer.SyncCall.Receives.EC2Client).To(Equal(ec2Client))
				Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(unsupported.KeyPair{}))
			})
		})

		Describe("cloud configurator", func() {
			It("generates a cloud config", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}

				_, err := command.Execute(globalFlags, incomingState)

				Expect(err).NotTo(HaveOccurred())
				Expect(cloudConfigurator.ConfigureCall.CallCount).To(Equal(1))
				Expect(cloudConfigurator.ConfigureCall.Receives.Stack).To(Equal(cloudformation.Stack{
					Name: "concourse",
				}))
				Expect(cloudConfigurator.ConfigureCall.Receives.AZs).To(ConsistOf("some-retrieved-az"))
			})
		})

		Describe("state manipulation", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					KeyPair: &storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
					},
					BOSH: &storage.BOSH{
						DirectorUsername:       "some-director-username",
						DirectorPassword:       "some-director-password",
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
						State: map[string]interface{}{
							"key": "value",
						},
					},
				}

				keyPairSynchronizer.SyncCall.Returns.KeyPair = unsupported.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				}
			})

			Context("aws keypair", func() {
				Context("when the keypair exists", func() {
					It("returns the given state unmodified", func() {
						state, err := command.Execute(globalFlags, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state).To(Equal(incomingState))
					})
				})

				Context("when the keypair doesn't exist", func() {
					It("returns the state with a new key pair", func() {
						incomingState.KeyPair = nil

						state, err := command.Execute(globalFlags, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(state.KeyPair).To(Equal(&storage.KeyPair{
							Name:       "some-keypair-name",
							PrivateKey: "some-private-key",
							PublicKey:  "some-public-key",
						}))
					})
				})
			})

			Describe("bosh", func() {
				Context("when the bosh director ssl keypair exists", func() {
					It("returns the given state unmodified", func() {
						state, err := command.Execute(globalFlags, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state).To(Equal(incomingState))
					})
				})

				Context("when the bosh director ssl keypair doesn't exist", func() {
					It("returns the state with a new key pair", func() {
						incomingState.BOSH = nil

						state, err := command.Execute(globalFlags, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.BOSH.DirectorSSLCertificate).To(Equal("updated-certificate"))
						Expect(state.BOSH.DirectorSSLPrivateKey).To(Equal("updated-private-key"))
						Expect(state.BOSH.State).To(Equal(map[string]interface{}{
							"updated-key": "updated-value",
						}))
					})
				})

				Context("when there are no director credentials", func() {
					It("deploys with randomized director credentials", func() {
						incomingState.BOSH = nil
						state, err := command.Execute(globalFlags, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(boshDeployer.DeployCall.Receives.Input.DirectorUsername).To(Equal("user-some-random-string"))
						Expect(boshDeployer.DeployCall.Receives.Input.DirectorPassword).To(Equal("p-some-random-string"))
						Expect(state.BOSH.DirectorPassword).To(Equal("p-some-random-string"))
					})
				})

				Context("when there are director credentials", func() {
					It("uses the old credentials", func() {
						_, err := command.Execute(globalFlags, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(boshDeployer.DeployCall.Receives.Input.DirectorUsername).To(Equal("some-director-username"))
						Expect(boshDeployer.DeployCall.Receives.Input.DirectorPassword).To(Equal("some-director-password"))
					})
				})

				Context("when the bosh credentials don't exist", func() {
					It("returns the state with random credentials", func() {
						incomingState.BOSH = nil
						boshDeployer.DeployCall.Returns.Output = unsupported.BOSHDeployOutput{
							Credentials: boshinit.InternalCredentials{
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

						state, err := command.Execute(globalFlags, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.BOSH.Credentials).To(Equal(boshinit.InternalCredentials{
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
						}))
					})

					Context("when the bosh credentials exist in the state.json", func() {
						It("deploys with those credentials and returns the state with the same credentials", func() {
							incomingState.BOSH = &storage.BOSH{
								Credentials: boshinit.InternalCredentials{
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

							state, err := command.Execute(globalFlags, incomingState)
							Expect(err).NotTo(HaveOccurred())
							Expect(boshDeployer.DeployCall.Receives.Input.Credentials).To(Equal(boshinit.InternalCredentials{
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
							}))
							Expect(state.BOSH.Credentials).To(Equal(boshinit.InternalCredentials{
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
							}))
						})
					})
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the cloudformation client can not be created", func() {
				clientProvider.CloudFormationClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error creating client"))
			})

			It("returns an error when the ec2 client can not be created", func() {
				clientProvider.EC2ClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error creating client"))
			})

			It("returns an error when the key pair fails to sync", func() {
				keyPairSynchronizer.SyncCall.Returns.Error = errors.New("error syncing key pair")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error syncing key pair"))
			})

			It("returns an error when infrastructure cannot be created", func() {
				infrastructureCreator.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("infrastructure creation failed"))
			})

			It("returns an error when the cloud config cannot be configured", func() {
				cloudConfigurator.ConfigureCall.Returns.Error = errors.New("bosh cloud configuration failed")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("bosh cloud configuration failed"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshDeployer := &fakes.BOSHDeployer{}
				boshDeployer.DeployCall.Returns.Error = errors.New("cannot deploy bosh")
				command = unsupported.NewDeployBOSHOnAWSForConcourse(
					infrastructureCreator, keyPairSynchronizer, clientProvider, boshDeployer,
					stringGenerator, cloudConfigurator, availabilityZoneRetriever)

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("cannot deploy bosh"))
			})

			It("returns an error when it cannot generate a string for the bosh director credentials", func() {
				stringGenerator.GenerateCall.Stub = nil
				stringGenerator.GenerateCall.Returns.Error = errors.New("cannot generate string")
				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("cannot generate string"))
			})

			It("returns an error when availability zones cannot be retrieved", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("availability zone could not be retrieved")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("availability zone could not be retrieved"))
			})
		})
	})
})
