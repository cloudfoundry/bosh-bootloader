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
		stateStore          *fakes.StateStore
		configurationParser application.ConfigurationParser
	)
	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		configurationParser = application.NewConfigurationParser(stateStore)
	})

	Describe("Parse", func() {
		It("returns a configuration with correct global flags based on arguments passed in", func() {
			args := []string{
				"--endpoint-override", "some-endpoint-override",
				"--aws-access-key-id", "some-access-key-id",
				"--aws-secret-access-key", "some-secret-access-key",
				"--aws-region", "some-region",
				"--state-dir", "some/state/dir",
				"some-command",
				"--subcommand-flag", "some-value",
			}
			configuration, err := configurationParser.Parse(args)
			Expect(err).NotTo(HaveOccurred())

			Expect(configuration.Global).To(Equal(application.GlobalConfiguration{
				Help:               false,
				Version:            false,
				EndpointOverride:   "some-endpoint-override",
				AWSAccessKeyID:     "some-access-key-id",
				AWSSecretAccessKey: "some-secret-access-key",
				AWSRegion:          "some-region",
				StateDir:           "some/state/dir",
			}))
		})

		It("returns a configuration with correct command with subcommand flags based on arguments passed in", func() {
			args := []string{
				"some-command",
				"--subcommand-flag", "some-value",
			}
			configuration, err := configurationParser.Parse(args)
			Expect(err).NotTo(HaveOccurred())

			Expect(configuration.Command).To(Equal("some-command"))
			Expect(configuration.SubcommandFlags).To(Equal([]string{"--subcommand-flag", "some-value"}))
		})

		Describe("state management", func() {
			It("returns a configuration with the state from the state store", func() {
				stateStore.GetCall.Returns.State = storage.State{
					Version: 1,
				}
				args := []string{
					"--state-dir", "some/state/dir",
					"some-command",
				}
				configuration, err := configurationParser.Parse(args)
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.GetCall.Receives.Dir).To(Equal("some/state/dir"))
				Expect(configuration.State).To(Equal(storage.State{
					Version: 1,
				}))
			})

			It("overrides aws configuration in the state with global flags", func() {
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

			Context("when no --state-dir is provided", func() {
				BeforeEach(func() {
					application.SetGetwd(func() (string, error) {
						return "some/state/dir", nil
					})
				})

				AfterEach(func() {
					application.ResetGetwd()
				})

				It("uses the current working directory as the state directory", func() {
					configuration, err := configurationParser.Parse([]string{
						"some-command",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(configuration.Global.StateDir).To(Equal("some/state/dir"))
				})
			})
		})

		Describe("command overrides", func() {
			DescribeTable("when a command is requested using a flag", func(commandLineArgument string, desiredCommand string) {
				configuration, err := configurationParser.Parse([]string{
					commandLineArgument,
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(configuration.Command).To(Equal(desiredCommand))
			},
				Entry("returns the help command provided --help", "--help", "help"),
				Entry("returns the help command provided --h", "--h", "help"),
				Entry("returns the help command provided help", "help", "help"),

				Entry("returns the version command provided --version", "--version", "version"),
				Entry("returns the version command provided --v", "--v", "version"),
				Entry("returns the version command provided version", "version", "version"),
			)
		})

		Context("failure cases", func() {
			It("returns a custom error when it fails to parse flags", func() {
				_, err := configurationParser.Parse([]string{
					"--invalid-flag",
					"some-command",
				})

				Expect(err).To(Equal(application.NewInvalidFlagError(
					errors.New("flag provided but not defined: -invalid-flag"),
				)))
			})

			It("returns an error when the command is not passed in", func() {
				_, err := configurationParser.Parse([]string{})
				Expect(err).To(MatchError(application.NewCommandNotProvidedError()))
			})

			It("returns an error when the state cannot be read", func() {
				stateStore.GetCall.Returns.Error = errors.New("failed to read state")
				_, err := configurationParser.Parse([]string{"some-command"})

				Expect(err).To(MatchError("failed to read state"))
			})

			It("returns an error when it cannot get working directory", func() {
				application.SetGetwd(func() (string, error) {
					return "", errors.New("failed to get working directory")
				})
				defer application.ResetGetwd()

				_, err := configurationParser.Parse([]string{
					"some-command",
				})
				Expect(err).To(MatchError("failed to get working directory"))
			})
		})
	})
})
