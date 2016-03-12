package unsupported_test

import (
	"errors"

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
			command               unsupported.DeployBOSHOnAWSForConcourse
			boshDeployer          *fakes.BOSHDeployer
			infrastructureCreator *fakes.InfrastructureCreator
			keyPairSynchronizer   *fakes.KeyPairSynchronizer
			cloudFormationClient  *fakes.CloudFormationClient
			ec2Client             *fakes.EC2Client
			clientProvider        *fakes.ClientProvider
			incomingState         storage.State
			globalFlags           commands.GlobalFlags
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
			boshDeployer.DeployCall.Returns.DirectorSSLKeyPair = ssl.KeyPair{
				Certificate: []byte("updated-certificate"),
				PrivateKey:  []byte("updated-private-key"),
			}
			boshDeployer.DeployCall.Returns.BOSHInitState = boshinit.State{
				"updated-key": "updated-value",
			}

			cloudFormationClient = &fakes.CloudFormationClient{}
			ec2Client = &fakes.EC2Client{}

			clientProvider = &fakes.ClientProvider{}
			clientProvider.CloudFormationClientCall.Returns.Client = cloudFormationClient
			clientProvider.EC2ClientCall.Returns.Client = ec2Client

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

			command = unsupported.NewDeployBOSHOnAWSForConcourse(infrastructureCreator, keyPairSynchronizer, clientProvider, boshDeployer)
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

			Context("bosh", func() {
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
							DirectorSSLCertificate: "updated-certificate",
							DirectorSSLPrivateKey:  "updated-private-key",
							State: map[string]interface{}{
								"updated-key": "updated-value",
							},
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

			It("returns an error when bosh cannot be deployed", func() {
				boshDeployer := &fakes.BOSHDeployer{}
				boshDeployer.DeployCall.Returns.Error = errors.New("cannot deploy bosh")
				command = unsupported.NewDeployBOSHOnAWSForConcourse(infrastructureCreator, keyPairSynchronizer, clientProvider, boshDeployer)

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("cannot deploy bosh"))
			})
		})
	})
})
