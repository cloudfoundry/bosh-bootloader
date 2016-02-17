package unsupported_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

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
			Expect(stateStore.MergeCall.Receives.Dir).To(Equal("/some/state/dir"))
			Expect(stateStore.MergeCall.Receives.Map).To(Equal(map[string]interface{}{
				"aws-access-key-id":     "some-aws-access-key-id",
				"aws-secret-access-key": "some-aws-secret-access-key",
				"aws-region":            "some-aws-region",
			}))
		})

		Context("when the aws access key id is not provided", func() {
			It("uses the AWS credentials from the state store if none are provided", func() {
				stateStore.GetStringCall.Returns.OK = true
				stateStore.GetStringCall.Returns.Value = "some-aws-access-key-id"

				err := command.Execute(commands.GlobalFlags{
					AWSSecretAccessKey: "some-aws-secret-access-key",
					AWSRegion:          "some-aws-region",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stateStore.GetStringCall.Receives.Dir).To(Equal("/some/state/dir"))
				Expect(stateStore.GetStringCall.Receives.Key).To(Equal("aws-access-key-id"))
				Expect(stateStore.MergeCall.Receives.Map).To(Equal(map[string]interface{}{
					"aws-access-key-id":     "some-aws-access-key-id",
					"aws-secret-access-key": "some-aws-secret-access-key",
					"aws-region":            "some-aws-region",
				}))
			})

			Context("when the state store fails", func() {
				It("returns an error", func() {
					stateStore.GetStringCall.Returns.Error = errors.New("get string failed")

					err := command.Execute(commands.GlobalFlags{
						AWSSecretAccessKey: "some-aws-secret-access-key",
						AWSRegion:          "some-aws-region",
						EndpointOverride:   "some-endpoint-override",
						StateDir:           "/some/state/dir",
					})
					Expect(err).To(MatchError("get string failed"))
				})
			})
		})

		Context("when the aws secret access key is not provided", func() {
			It("uses the AWS credentials from the state store if none are provided", func() {
				stateStore.GetStringCall.Returns.OK = true
				stateStore.GetStringCall.Returns.Value = "some-aws-secret-access-key"

				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:   "some-aws-access-key-id",
					AWSRegion:        "some-aws-region",
					EndpointOverride: "some-endpoint-override",
					StateDir:         "/some/state/dir",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stateStore.GetStringCall.Receives.Dir).To(Equal("/some/state/dir"))
				Expect(stateStore.GetStringCall.Receives.Key).To(Equal("aws-secret-access-key"))
				Expect(stateStore.MergeCall.Receives.Map).To(Equal(map[string]interface{}{
					"aws-access-key-id":     "some-aws-access-key-id",
					"aws-secret-access-key": "some-aws-secret-access-key",
					"aws-region":            "some-aws-region",
				}))
			})

			Context("when the state store fails", func() {
				It("returns an error", func() {
					stateStore.GetStringCall.Returns.Error = errors.New("get string failed")

					err := command.Execute(commands.GlobalFlags{
						AWSAccessKeyID:   "some-aws-access-key-id",
						AWSRegion:        "some-aws-region",
						EndpointOverride: "some-endpoint-override",
						StateDir:         "/some/state/dir",
					})
					Expect(err).To(MatchError("get string failed"))
				})
			})
		})

		Context("when the aws region is not provided", func() {
			It("uses the AWS credentials from the state store if none are provided", func() {
				stateStore.GetStringCall.Returns.OK = true
				stateStore.GetStringCall.Returns.Value = "some-aws-region"

				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:     "some-aws-access-key-id",
					AWSSecretAccessKey: "some-aws-secret-access-key",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stateStore.GetStringCall.Receives.Dir).To(Equal("/some/state/dir"))
				Expect(stateStore.GetStringCall.Receives.Key).To(Equal("aws-region"))
				Expect(stateStore.MergeCall.Receives.Map).To(Equal(map[string]interface{}{
					"aws-access-key-id":     "some-aws-access-key-id",
					"aws-secret-access-key": "some-aws-secret-access-key",
					"aws-region":            "some-aws-region",
				}))
			})

			Context("when the state store fails", func() {
				It("returns an error", func() {
					stateStore.GetStringCall.Returns.Error = errors.New("get string failed")

					err := command.Execute(commands.GlobalFlags{
						AWSAccessKeyID:     "some-aws-access-key-id",
						AWSSecretAccessKey: "some-aws-secret-access-key",
						EndpointOverride:   "some-endpoint-override",
						StateDir:           "/some/state/dir",
					})
					Expect(err).To(MatchError("get string failed"))
				})
			})
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
				stateStore.MergeCall.Returns.Error = errors.New("state store merge failed")

				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:     "some-aws-access-key-id",
					AWSSecretAccessKey: "some-aws-secret-access-key",
					AWSRegion:          "some-aws-region",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).To(MatchError("state store merge failed"))
			})

			It("returns an error when the access key id is missing", func() {
				err := command.Execute(commands.GlobalFlags{
					AWSSecretAccessKey: "some-aws-secret-access-key",
					AWSRegion:          "some-aws-region",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).To(MatchError("aws credentials must be provided"))
			})

			It("returns an error when the secret access key is missing", func() {
				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:   "some-aws-access-key-id",
					AWSRegion:        "some-aws-region",
					EndpointOverride: "some-endpoint-override",
					StateDir:         "/some/state/dir",
				})
				Expect(err).To(MatchError("aws credentials must be provided"))
			})

			It("returns an error when the region is missing", func() {
				err := command.Execute(commands.GlobalFlags{
					AWSAccessKeyID:     "some-aws-access-key-id",
					AWSSecretAccessKey: "some-aws-secret-access-key",
					EndpointOverride:   "some-endpoint-override",
					StateDir:           "/some/state/dir",
				})
				Expect(err).To(MatchError("aws credentials must be provided"))
			})
		})
	})
})
