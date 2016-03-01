package unsupported_test

import (
	"errors"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProvisionAWSForConcourse", func() {
	Describe("Execute", func() {
		var (
			command                      unsupported.ProvisionAWSForConcourse
			builder                      *fakes.TemplateBuilder
			stackManager                 *fakes.StackManager
			keyPairManager               *fakes.KeyPairManager
			cloudFormationClient         *fakes.CloudFormationClient
			cloudFormationClientProvider *fakes.CloudFormationClientProvider
			ec2Client                    *fakes.EC2Client
			ec2ClientProvider            *fakes.EC2ClientProvider
			incomingState                storage.State
			globalFlags                  commands.GlobalFlags
		)

		BeforeEach(func() {
			builder = &fakes.TemplateBuilder{}

			cloudFormationClient = &fakes.CloudFormationClient{}
			cloudFormationClientProvider = &fakes.CloudFormationClientProvider{}
			cloudFormationClientProvider.ClientCall.Returns.Client = cloudFormationClient

			ec2Client = &fakes.EC2Client{}
			ec2ClientProvider = &fakes.EC2ClientProvider{}
			ec2ClientProvider.ClientCall.Returns.Client = ec2Client

			stackManager = &fakes.StackManager{}
			keyPairManager = &fakes.KeyPairManager{}

			command = unsupported.NewProvisionAWSForConcourse(builder, stackManager, keyPairManager, cloudFormationClientProvider, ec2ClientProvider)

			builder.BuildCall.Returns.Template = cloudformation.Template{
				AWSTemplateFormatVersion: "some-template-version",
				Description:              "some-description",
				Parameters: map[string]cloudformation.Parameter{
					"KeyName": {
						Type:        "AWS::EC2::KeyPair::KeyName",
						Default:     "some-keypair-name",
						Description: "SSH KeyPair to use for instances",
					},
				},
				Mappings:  map[string]interface{}{},
				Resources: map[string]cloudformation.Resource{},
			}

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
			}

			keyPairManager.SyncCall.Returns.KeyPair = ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: []byte("some-private-key"),
				PublicKey:  []byte("some-public-key"),
			}
		})

		It("creates/updates the stack with the given name", func() {
			_, err := command.Execute(globalFlags, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClientProvider.ClientCall.Receives.Config).To(Equal(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint",
			}))
			Expect(builder.BuildCall.Receives.KeyPairName).To(Equal("some-keypair-name"))
			Expect(stackManager.CreateOrUpdateCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.CreateOrUpdateCall.Receives.StackName).To(Equal("concourse"))
			Expect(stackManager.CreateOrUpdateCall.Receives.Template).To(Equal(cloudformation.Template{
				AWSTemplateFormatVersion: "some-template-version",
				Description:              "some-description",
				Parameters: map[string]cloudformation.Parameter{
					"KeyName": {
						Type:        "AWS::EC2::KeyPair::KeyName",
						Default:     "some-keypair-name",
						Description: "SSH KeyPair to use for instances",
					},
				},
				Mappings:  map[string]interface{}{},
				Resources: map[string]cloudformation.Resource{},
			}))

			Expect(stackManager.WaitForCompletionCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.WaitForCompletionCall.Receives.StackName).To(Equal("concourse"))
			Expect(stackManager.WaitForCompletionCall.Receives.SleepInterval).To(Equal(2 * time.Second))
		})

		It("syncs the keypair", func() {
			state, err := command.Execute(globalFlags, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(ec2ClientProvider.ClientCall.Receives.Config).To(Equal(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint",
			}))
			Expect(keyPairManager.SyncCall.Receives.EC2Client).To(Equal(ec2Client))
			Expect(keyPairManager.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: []byte("some-private-key"),
				PublicKey:  []byte("some-public-key"),
			}))

			Expect(state.KeyPair).To(Equal(&storage.KeyPair{
				Name:       "some-keypair-name",
				PublicKey:  "some-public-key",
				PrivateKey: "some-private-key",
			}))
		})

		It("returns the given state unmodified", func() {
			_, err := command.Execute(globalFlags, incomingState)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is no keypair", func() {
			BeforeEach(func() {
				incomingState.KeyPair = nil
			})

			It("syncs with an empty keypair", func() {
				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairManager.SyncCall.Receives.EC2Client).To(Equal(ec2Client))
				Expect(keyPairManager.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
					Name:       "",
					PrivateKey: []byte(""),
					PublicKey:  []byte(""),
				}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the cloudformation client can not be created", func() {
				cloudFormationClientProvider.ClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error creating client"))
			})

			It("returns an error when the ec2 client can not be created", func() {
				ec2ClientProvider.ClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error creating client"))
			})

			It("returns an error when the key pair fails to sync", func() {
				keyPairManager.SyncCall.Returns.Error = errors.New("error syncing key pair")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error syncing key pair"))
			})

			It("returns an error when the stack can not be created", func() {
				stackManager.CreateOrUpdateCall.Returns.Error = errors.New("error creating stack")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error creating stack"))
			})

			It("returns an error when waiting for completion errors", func() {
				stackManager.WaitForCompletionCall.Returns.Error = errors.New("error waiting on stack")

				_, err := command.Execute(globalFlags, incomingState)
				Expect(err).To(MatchError("error waiting on stack"))
			})
		})
	})
})
