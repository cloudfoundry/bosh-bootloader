package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

var _ = Describe("ConfigurationParser", func() {
	var (
		stateStore             *fakes.StateStore
		commandLineParser      *fakes.CommandLineParser
		awsCredentialValidator *fakes.AWSCredentialValidator
		configurationParser    application.ConfigurationParser
	)
	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		commandLineParser = &fakes.CommandLineParser{}
		awsCredentialValidator = &fakes.AWSCredentialValidator{}
		configurationParser = application.NewConfigurationParser(commandLineParser, awsCredentialValidator, stateStore)
	})

	Describe("Parse", func() {
		It("returns a configuration based on arguments provided", func() {
			commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
				AWSAccessKeyID:     "access-key-id-from-flag",
				AWSSecretAccessKey: "secret-access-key-from-flag",
				AWSRegion:          "region-from-flag",
				Command:            "unsupported-deploy-bosh-on-aws-for-concourse",
				SubcommandFlags:    []string{"--some-flag", "some-value"},
				StateDir:           "some/state/dir",
				EndpointOverride:   "some-endpoint-override",
			}
			configuration, err := configurationParser.Parse([]string{"unsupported-deploy-bosh-on-aws-for-concourse"})
			Expect(err).NotTo(HaveOccurred())

			Expect(configuration.Command).To(Equal("unsupported-deploy-bosh-on-aws-for-concourse"))
			Expect(configuration.SubcommandFlags).To(Equal([]string{"--some-flag", "some-value"}))
			Expect(configuration.Global).To(Equal(application.GlobalConfiguration{
				EndpointOverride: "some-endpoint-override",
				StateDir:         "some/state/dir",
			}))

			Expect(commandLineParser.ParseCall.Receives.Arguments).To(Equal([]string{"unsupported-deploy-bosh-on-aws-for-concourse"}))
		})

		Describe("command validation", func() {
			Context("when an unknown command is provided", func() {
				It("returns an error", func() {
					commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
						Command: "unknown-command",
					}
					_, err := configurationParser.Parse([]string{"unknown-command"})
					Expect(err).To(Equal(application.NewInvalidCommandError(errors.New("unknown command: unknown-command"))))
				})
			})

			Context("when nothing is provided", func() {
				It("returns an error", func() {
					commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
						Command: "",
					}
					_, err := configurationParser.Parse([]string{})
					Expect(err).To(Equal(application.NewInvalidCommandError(errors.New("unknown command: [EMPTY]"))))
				})
			})
		})

		Describe("credential validation", func() {
			DescribeTable("when credential validation is not required", func(command string) {
				commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
					Command: command,
				}
				_, err := configurationParser.Parse([]string{command})
				Expect(err).NotTo(HaveOccurred())

				Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(0))
			},
				Entry("does not validate credentials for help", "help"),
				Entry("does not validate credentials for version", "version"),
				Entry("does not validate credentials for director-address", "director-address"),
				Entry("does not validate credentials for director-username", "director-username"),
				Entry("does not validate credentials for director-password", "director-password"),
				Entry("does not validate credentials for ssh-key", "ssh-key"),
			)

			DescribeTable("when credential validation is required", func(command string) {
				commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
					AWSAccessKeyID:     "some-access-key-id",
					AWSSecretAccessKey: "some-secret-access-key",
					AWSRegion:          "some-region",
					Command:            command,
				}
				awsCredentialValidator.ValidateCall.Returns.Error = errors.New("credential validation failed")

				_, err := configurationParser.Parse([]string{command})
				Expect(err).To(MatchError("credential validation failed"))

				Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(awsCredentialValidator.ValidateCall.Receives.AccessKeyID).To(Equal("some-access-key-id"))
				Expect(awsCredentialValidator.ValidateCall.Receives.SecretAccessKey).To(Equal("some-secret-access-key"))
				Expect(awsCredentialValidator.ValidateCall.Receives.Region).To(Equal("some-region"))
			},
				Entry("validates credentials for unsupported-deploy-bosh-on-aws-for-concourse", "unsupported-deploy-bosh-on-aws-for-concourse"),
				Entry("validates credentials for destroy", "destroy"),
				Entry("validates credentials for unsupported-create-lbs", "unsupported-create-lbs"),
				Entry("validates credentials for unsupported-update-lbs", "unsupported-update-lbs"),
			)
		})

		Describe("state management", func() {
			It("returns a configuration with the state from the state store", func() {
				stateStore.GetCall.Returns.State = storage.State{
					Version: 1,
				}
				commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
					AWSAccessKeyID:     "access-key-id-from-flag",
					AWSSecretAccessKey: "secret-access-key-from-flag",
					AWSRegion:          "region-from-flag",
					StateDir:           "some/state/dir",
					Command:            "help",
				}
				configuration, err := configurationParser.Parse([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.GetCall.Receives.Dir).To(Equal("some/state/dir"))
				Expect(configuration.State).To(Equal(storage.State{
					Version: 1,
					AWS: storage.AWS{
						AccessKeyID:     "access-key-id-from-flag",
						SecretAccessKey: "secret-access-key-from-flag",
						Region:          "region-from-flag",
					},
				}))
			})

			It("overrides aws configuration in the state with global flags", func() {
				commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
					AWSAccessKeyID:     "access-key-id-from-flag",
					AWSSecretAccessKey: "secret-access-key-from-flag",
					AWSRegion:          "region-from-flag",
					Command:            "unsupported-deploy-bosh-on-aws-for-concourse",
				}

				stateStore.GetCall.Returns.State = storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "access-key-id-from-state",
						SecretAccessKey: "secret-access-key-from-state",
						Region:          "region-from-state",
					},
				}

				configuration, err := configurationParser.Parse([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(configuration.State.AWS).To(Equal(storage.AWS{
					AccessKeyID:     "access-key-id-from-flag",
					SecretAccessKey: "secret-access-key-from-flag",
					Region:          "region-from-flag",
				}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the command line cannot be parsed", func() {
				commandLineParser.ParseCall.Returns.Error = errors.New("failed to parse command line")
				_, err := configurationParser.Parse([]string{"some-command"})

				Expect(err).To(MatchError("failed to parse command line"))
			})

			It("returns an error when the state cannot be read", func() {
				stateStore.GetCall.Returns.Error = errors.New("failed to read state")
				_, err := configurationParser.Parse([]string{"some-command"})

				Expect(err).To(MatchError("failed to read state"))
			})
		})
	})
})
