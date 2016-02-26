package unsupported_test

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
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
			command         unsupported.ProvisionAWSForConcourse
			builder         *fakes.TemplateBuilder
			manager         *fakes.StackManager
			session         *fakes.CloudFormationSession
			sessionProvider *fakes.CloudFormationSessionProvider
			incomingState   storage.State
		)

		BeforeEach(func() {
			builder = &fakes.TemplateBuilder{}

			session = &fakes.CloudFormationSession{}
			sessionProvider = &fakes.CloudFormationSessionProvider{}
			sessionProvider.SessionCall.Returns.Session = session

			manager = &fakes.StackManager{}

			command = unsupported.NewProvisionAWSForConcourse(builder, manager, sessionProvider)

			builder.BuildCall.Returns.Template = cloudformation.Template{
				AWSTemplateFormatVersion: "some-template-version",
				Description:              "some-description",
				Parameters:               map[string]cloudformation.Parameter{},
				Mappings:                 map[string]interface{}{},
				Resources:                map[string]cloudformation.Resource{},
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

		})

		It("creates/updates the stack with the given name", func() {
			_, err := command.Execute(commands.GlobalFlags{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(manager.CreateOrUpdateCall.Receives.Session).To(Equal(session))
			Expect(manager.CreateOrUpdateCall.Receives.StackName).To(Equal("concourse"))

			buf, err := json.Marshal(manager.CreateOrUpdateCall.Receives.Template)
			Expect(err).NotTo(HaveOccurred())
			Expect(buf).To(MatchJSON(`{
				"AWSTemplateFormatVersion": "some-template-version",
				"Description": "some-description",
				"Parameters": {
					"KeyName": {
						"Type":        "AWS::EC2::KeyPair::KeyName",
						"Default":     "some-keypair-name",
						"Description": "SSH Keypair to use for instances"
					}
				}
			}`))

			Expect(manager.WaitForCompletionCall.Receives.Session).To(Equal(session))
			Expect(manager.WaitForCompletionCall.Receives.StackName).To(Equal("concourse"))
			Expect(manager.WaitForCompletionCall.Receives.SleepInterval).To(Equal(2 * time.Second))
		})

		It("returns the given state unmodified", func() {
			_, err := command.Execute(commands.GlobalFlags{}, incomingState)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is no keypair", func() {
			It("returns an error when a keypair does not exist", func() {
				_, err := command.Execute(commands.GlobalFlags{}, storage.State{})
				Expect(err).To(MatchError("no keypair is present, you can generate a keypair by running the unsupported-create-bosh-aws-keypair command."))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the session can not be created", func() {
				sessionProvider.SessionCall.Returns.Error = errors.New("error creating session")

				_, err := command.Execute(commands.GlobalFlags{}, incomingState)
				Expect(err).To(MatchError("error creating session"))
			})

			It("returns an error when the stack can not be created", func() {
				manager.CreateOrUpdateCall.Returns.Error = errors.New("error creating stack")

				_, err := command.Execute(commands.GlobalFlags{}, incomingState)
				Expect(err).To(MatchError("error creating stack"))
			})

			It("returns an error when waiting for completion errors", func() {
				manager.WaitForCompletionCall.Returns.Error = errors.New("error waiting on stack")

				_, err := command.Execute(commands.GlobalFlags{}, incomingState)
				Expect(err).To(MatchError("error waiting on stack"))
			})
		})
	})
})
