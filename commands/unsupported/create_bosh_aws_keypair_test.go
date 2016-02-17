package unsupported_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateBoshAWSKeypair", func() {
	var (
		command          unsupported.CreateBoshAWSKeypair
		keypairGenerator *fakes.KeypairGenerator
		keypairUploader  *fakes.KeypairUploader
		stateStore       *fakes.StateStore
		session          *fakes.Session
		sessionProvider  *fakes.SessionProvider
	)

	BeforeEach(func() {
		keypairGenerator = &fakes.KeypairGenerator{}
		keypairUploader = &fakes.KeypairUploader{}
		stateStore = &fakes.StateStore{}

		session = &fakes.Session{}
		sessionProvider = &fakes.SessionProvider{}
		sessionProvider.SessionCall.Returns.Session = session

		command = unsupported.NewCreateBoshAWSKeypair(keypairGenerator, keypairUploader, sessionProvider, stateStore)
	})

	Describe("Execute", func() {
		It("generates a new keypair", func() {
			err := command.Execute(commands.GlobalFlags{
				AWSAccessKeyID:     "some-aws-access-key-id",
				AWSSecretAccessKey: "some-aws-secret-access-key",
				AWSRegion:          "some-aws-region",
				EndpointOverride:   "some-endpoint-override",
				StateDir:           "/some/state/dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(keypairGenerator.GenerateCall.CallCount).To(Equal(1))
		})

		It("initializes a new session with the correct config", func() {
			err := command.Execute(commands.GlobalFlags{
				AWSAccessKeyID:     "some-aws-access-key-id",
				AWSSecretAccessKey: "some-aws-secret-access-key",
				AWSRegion:          "some-aws-region",
				EndpointOverride:   "some-endpoint-override",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(sessionProvider.SessionCall.Receives.Config).To(Equal(ec2.Config{
				AccessKeyID:      "some-aws-access-key-id",
				SecretAccessKey:  "some-aws-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint-override",
			}))
		})

		It("uploads the generated keypair", func() {
			keypairGenerator.GenerateCall.Returns.Keypair = ec2.Keypair{
				Name: "some-name",
				Key:  []byte("some-key"),
			}

			err := command.Execute(commands.GlobalFlags{
				AWSAccessKeyID:     "some-aws-access-key-id",
				AWSSecretAccessKey: "some-aws-secret-access-key",
				AWSRegion:          "some-aws-region",
				EndpointOverride:   "some-endpoint-override",
				StateDir:           "/some/state/dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(keypairUploader.UploadCall.Receives.Session).To(Equal(session))
			Expect(keypairUploader.UploadCall.Receives.Keypair).To(Equal(ec2.Keypair{
				Name: "some-name",
				Key:  []byte("some-key"),
			}))
		})

		It("stores the AWS credentials", func() {
			err := command.Execute(commands.GlobalFlags{
				AWSAccessKeyID:     "some-aws-access-key-id",
				AWSSecretAccessKey: "some-aws-secret-access-key",
				AWSRegion:          "some-aws-region",
				EndpointOverride:   "some-endpoint-override",
				StateDir:           "/some/state/dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stateStore.SetCall.Receives.Dir).To(Equal("/some/state/dir"))
			Expect(stateStore.SetCall.Receives.State).To(Equal(state.State{
				AWSAccessKeyID:     "some-aws-access-key-id",
				AWSSecretAccessKey: "some-aws-secret-access-key",
				AWSRegion:          "some-aws-region",
			}))
		})

		Context("failure cases", func() {
			It("returns an error when key generation fails", func() {
				keypairGenerator.GenerateCall.Returns.Error = errors.New("generate keys failed")

				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:     "some-aws-access-key-id",
					AWSSecretAccessKey: "some-aws-secret-access-key",
					AWSRegion:          "some-aws-region",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).To(MatchError("generate keys failed"))
			})

			It("returns an error when key upload fails", func() {
				keypairUploader.UploadCall.Returns.Error = errors.New("upload keys failed")

				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:     "some-aws-access-key-id",
					AWSSecretAccessKey: "some-aws-secret-access-key",
					AWSRegion:          "some-aws-region",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).To(MatchError("upload keys failed"))
			})

			It("returns an error when state store fails", func() {
				stateStore.SetCall.Returns.Error = errors.New("state store merge failed")

				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:     "some-aws-access-key-id",
					AWSSecretAccessKey: "some-aws-secret-access-key",
					AWSRegion:          "some-aws-region",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).To(MatchError("state store merge failed"))
			})

			It("returns an error when the session provided fails", func() {
				sessionProvider.SessionCall.Returns.Error = errors.New("failed to create session")

				err := command.Execute(commands.GlobalFlags{
					AWSSecretAccessKey: "some-aws-secret-access-key",
					AWSRegion:          "some-aws-region",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).To(MatchError("failed to create session"))
			})
		})
	})
})
