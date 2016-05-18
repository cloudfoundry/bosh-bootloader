package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

var _ = Describe("ConfigurationParser", func() {
	var (
		stateStore          *fakes.StateStore
		commandLineParser   *fakes.CommandLineParser
		configurationParser application.ConfigurationParser
	)
	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		commandLineParser = &fakes.CommandLineParser{}
		configurationParser = application.NewConfigurationParser(commandLineParser, stateStore)
	})

	Describe("Parse", func() {
		It("returns a configuration based on arguments provided", func() {
			commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
				Command:          "some-command",
				SubcommandFlags:  []string{"--some-flag", "some-value"},
				StateDir:         "some/state/dir",
				EndpointOverride: "some-endpoint-override",
			}
			configuration, err := configurationParser.Parse([]string{"some-command"})
			Expect(err).NotTo(HaveOccurred())

			Expect(configuration.Command).To(Equal("some-command"))
			Expect(configuration.SubcommandFlags).To(Equal([]string{"--some-flag", "some-value"}))
			Expect(configuration.Global).To(Equal(application.GlobalConfiguration{
				EndpointOverride: "some-endpoint-override",
				StateDir:         "some/state/dir",
			}))

			Expect(commandLineParser.ParseCall.Receives.Arguments).To(Equal([]string{"some-command"}))
		})

		Describe("state management", func() {
			It("returns a configuration with the state from the state store", func() {
				stateStore.GetCall.Returns.State = storage.State{
					Version: 1,
				}
				commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
					StateDir: "some/state/dir",
				}
				configuration, err := configurationParser.Parse([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.GetCall.Receives.Dir).To(Equal("some/state/dir"))
				Expect(configuration.State).To(Equal(storage.State{
					Version: 1,
				}))
			})

			It("overrides aws configuration in the state with global flags", func() {
				commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
					AWSAccessKeyID:     "access-key-id-from-flag",
					AWSSecretAccessKey: "secret-access-key-from-flag",
					AWSRegion:          "region-from-flag",
				}

				stateStore.GetCall.Returns.State = storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "access-key-id-from-state",
						SecretAccessKey: "secret-access-key-from-state",
						Region:          "region-from-state",
					},
				}

				configuration, err := configurationParser.Parse([]string{
					"--aws-access-key-id", "access-key-id-from-flag",
					"--aws-secret-access-key", "secret-access-key-from-flag",
					"--aws-region", "region-from-flag",
					"some-command",
				})
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
