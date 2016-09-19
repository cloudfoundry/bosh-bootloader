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
		commandLineParser   *fakes.CommandLineParser
		configurationParser application.ConfigurationParser
	)
	BeforeEach(func() {
		commandLineParser = &fakes.CommandLineParser{}
		configurationParser = application.NewConfigurationParser(commandLineParser)

		application.SetGetState(func(dir string) (storage.State, error) {
			return storage.State{Version: 1}, nil
		})
	})

	AfterEach(func() {
		application.ResetGetState()
	})

	Describe("Parse", func() {
		It("returns a configuration based on arguments provided", func() {

			commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
				AWSAccessKeyID:     "access-key-id-from-flag",
				AWSSecretAccessKey: "secret-access-key-from-flag",
				AWSRegion:          "region-from-flag",
				Command:            "up",
				SubcommandFlags:    []string{"--some-flag", "some-value"},
				StateDir:           "some/state/dir",
				EndpointOverride:   "some-endpoint-override",
			}
			configuration, err := configurationParser.Parse([]string{"up"})
			Expect(err).NotTo(HaveOccurred())

			Expect(configuration.Command).To(Equal("up"))
			Expect(configuration.SubcommandFlags).To(Equal(application.StringSlice{"--some-flag", "some-value"}))
			Expect(configuration.Global).To(Equal(application.GlobalConfiguration{
				EndpointOverride: "some-endpoint-override",
				StateDir:         "some/state/dir",
			}))

			Expect(commandLineParser.ParseCall.Receives.Arguments).To(Equal([]string{"up"}))
		})

		Describe("state management", func() {
			It("returns a configuration with the state from the state store", func() {
				commandLineParser.ParseCall.Returns.CommandLineConfiguration = application.CommandLineConfiguration{
					AWSAccessKeyID:     "access-key-id-from-flag",
					AWSSecretAccessKey: "secret-access-key-from-flag",
					AWSRegion:          "region-from-flag",
					StateDir:           "some/state/dir",
					Command:            "help",
				}
				configuration, err := configurationParser.Parse([]string{})
				Expect(err).NotTo(HaveOccurred())

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
					Command:            "up",
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
				application.SetGetState(func(dir string) (storage.State, error) {
					return storage.State{}, errors.New("failed to read state")
				})

				_, err := configurationParser.Parse([]string{"some-command"})

				Expect(err).To(MatchError("failed to read state"))
			})
		})
	})
})
