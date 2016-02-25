package unsupported_test

import (
	"encoding/json"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProvisionAWSForConcourse", func() {
	Describe("Execute", func() {
		var (
			command         unsupported.ProvisionAWSForConcourse
			builder         *fakes.TemplateBuilder
			creator         *fakes.StackCreator
			session         *fakes.CloudFormationSession
			sessionProvider *fakes.CloudFormationSessionProvider
			incomingState   state.State
		)

		BeforeEach(func() {
			builder = &fakes.TemplateBuilder{}

			session = &fakes.CloudFormationSession{}
			sessionProvider = &fakes.CloudFormationSessionProvider{}
			sessionProvider.SessionCall.Returns.Session = session

			creator = &fakes.StackCreator{}

			command = unsupported.NewProvisionAWSForConcourse(builder, creator, sessionProvider)

			builder.BuildCall.Returns.Template = cloudformation.Template{
				AWSTemplateFormatVersion: "some-template-version",
				Description:              "some-description",
				Parameters:               map[string]cloudformation.Parameter{},
				Mappings:                 map[string]interface{}{},
				Resources:                map[string]cloudformation.Resource{},
			}

			incomingState = state.State{
				AWS: state.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				KeyPair: &state.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}

		})

		It("creates a stack with the keypair given in the state dir", func() {
			_, err := command.Execute(commands.GlobalFlags{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(creator.CreateCall.Receives.Session).To(Equal(session))

			buf, err := json.MarshalIndent(creator.CreateCall.Receives.Template, "", "  ")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(buf)).To(MatchJSON(`{
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

		})

		It("returns the given state unmodified", func() {
			_, err := command.Execute(commands.GlobalFlags{}, incomingState)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is no keypair", func() {
			It("returns an error when a keypair does not exist", func() {
				_, err := command.Execute(commands.GlobalFlags{}, state.State{})
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
				creator.CreateCall.Returns.Error = errors.New("error creating stack")

				_, err := command.Execute(commands.GlobalFlags{}, incomingState)
				Expect(err).To(MatchError("error creating stack"))
			})
		})
	})
})
