package unsupported_test

import (
	"bytes"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
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
			command                 unsupported.DeployBOSHOnAWSForConcourse
			stdout                  *bytes.Buffer
			stackManager            *fakes.StackManager
			infrastructureCreator   *fakes.InfrastructureCreator
			keyPairSynchronizer     *fakes.KeyPairSynchronizer
			cloudFormationClient    *fakes.CloudFormationClient
			clientProvider          *fakes.ClientProvider
			ec2Client               *fakes.EC2Client
			logger                  *fakes.Logger
			boshInitManifestBuilder *fakes.BOSHInitManifestBuilder
			incomingState           storage.State
			globalFlags             commands.GlobalFlags
		)

		BeforeEach(func() {
			stdout = bytes.NewBuffer([]byte{})

			cloudFormationClient = &fakes.CloudFormationClient{}
			ec2Client = &fakes.EC2Client{}

			clientProvider = &fakes.ClientProvider{}
			clientProvider.CloudFormationClientCall.Returns.Client = cloudFormationClient
			clientProvider.EC2ClientCall.Returns.Client = ec2Client

			infrastructureCreator = &fakes.InfrastructureCreator{}
			stackManager = &fakes.StackManager{}

			keyPairSynchronizer = &fakes.KeyPairSynchronizer{}

			logger = &fakes.Logger{}

			boshInitManifestBuilder = &fakes.BOSHInitManifestBuilder{}
			boshInitManifestBuilder.BuildCall.Returns.Manifest = boshinit.Manifest{
				Name: "bosh",
				Jobs: []boshinit.Job{{
					Properties: boshinit.JobProperties{
						Director: boshinit.DirectorJobProperties{
							SSL: boshinit.SSLProperties{
								Cert: "some-certificate",
								Key:  "some-private-key",
							},
						},
					},
				}},
			}

			command = unsupported.NewDeployBOSHOnAWSForConcourse(stackManager, infrastructureCreator, keyPairSynchronizer, clientProvider, boshInitManifestBuilder, stdout)

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
				},
			}

			keyPairSynchronizer.SyncCall.Returns.KeyPair = unsupported.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}
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

		It("prints out the bosh-init manifest", func() {
			stackManager.DescribeCall.Returns.Output = cloudformation.Stack{
				Outputs: map[string]string{
					"BOSHSubnet":              "subnet-12345",
					"BOSHSubnetAZ":            "some-az",
					"BOSHEIP":                 "some-elastic-ip",
					"BOSHUserAccessKey":       "some-access-key-id",
					"BOSHUserSecretAccessKey": "some-secret-access-key",
					"BOSHSecurityGroup":       "some-security-group",
				},
			}

			_, err := command.Execute(globalFlags, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshInitManifestBuilder.BuildCall.Receives.Properties).To(Equal(boshinit.ManifestProperties{
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
			}))
			Expect(stdout.String()).To(ContainSubstring("bosh-init manifest:"))
			Expect(stdout.String()).To(ContainSubstring(`name: bosh`))
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

		Context("state manipulation", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					KeyPair: &storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
					},
					BOSH: &storage.BOSH{
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
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

			Context("ssl keypair", func() {
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
						Expect(state.BOSH).To(Equal(&storage.BOSH{
							DirectorSSLCertificate: "some-certificate",
							DirectorSSLPrivateKey:  "some-private-key",
						}))
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

			It("returns an error when describe stacks returns an error", func() {
				stackManager.DescribeCall.Returns.Error = errors.New("error describing stack")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error describing stack"))
			})

			It("returns an error when the bosh-init manifest cannot be built", func() {
				boshInitManifestBuilder := &fakes.BOSHInitManifestBuilder{}
				boshInitManifestBuilder.BuildCall.Returns.Error = errors.New("cannot build manifest")
				command = unsupported.NewDeployBOSHOnAWSForConcourse(stackManager, infrastructureCreator, keyPairSynchronizer, clientProvider, boshInitManifestBuilder, stdout)

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("cannot build manifest"))
			})
		})
	})
})
