package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
)

var _ = Describe("CommandLineParser", func() {
	var commandLineParser application.CommandLineParser

	BeforeEach(func() {
		commandLineParser = application.NewCommandLineParser()
	})

	Describe("Parse", func() {
		It("returns a command line configuration with correct global flags based on arguments passed in", func() {
			args := []string{
				"--endpoint-override", "some-endpoint-override",
				"--aws-access-key-id", "some-access-key-id",
				"--aws-secret-access-key", "some-secret-access-key",
				"--aws-region", "some-region",
				"--state-dir", "some/state/dir",
				"some-command",
				"--subcommand-flag", "some-value",
			}
			commandLineConfiguration, err := commandLineParser.Parse(args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandLineConfiguration.EndpointOverride).To(Equal("some-endpoint-override"))
			Expect(commandLineConfiguration.StateDir).To(Equal("some/state/dir"))
			Expect(commandLineConfiguration.AWSAccessKeyID).To(Equal("some-access-key-id"))
			Expect(commandLineConfiguration.AWSSecretAccessKey).To(Equal("some-secret-access-key"))
			Expect(commandLineConfiguration.AWSRegion).To(Equal("some-region"))
		})

		It("returns a command line configuration with correct command with subcommand flags based on arguments passed in", func() {
			args := []string{
				"some-command",
				"--subcommand-flag", "some-value",
			}
			commandLineConfiguration, err := commandLineParser.Parse(args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandLineConfiguration.Command).To(Equal("some-command"))
			Expect(commandLineConfiguration.SubcommandFlags).To(Equal([]string{"--subcommand-flag", "some-value"}))
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
				commandLineConfiguration, err := commandLineParser.Parse([]string{
					"some-command",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(commandLineConfiguration.StateDir).To(Equal("some/state/dir"))
			})
		})

		DescribeTable("when a command is requested using a flag", func(commandLineArgument string, desiredCommand string) {
			commandLineConfiguration, err := commandLineParser.Parse([]string{
				commandLineArgument,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(commandLineConfiguration.Command).To(Equal(desiredCommand))
		},
			Entry("returns the help command provided --help", "--help", "help"),
			Entry("returns the help command provided --h", "--h", "help"),
			Entry("returns the help command provided help", "help", "help"),

			Entry("returns the version command provided --version", "--version", "version"),
			Entry("returns the version command provided --v", "--v", "version"),
			Entry("returns the version command provided version", "version", "version"),
		)

		Context("failure cases", func() {
			It("returns a custom error when it fails to parse flags", func() {
				_, err := commandLineParser.Parse([]string{
					"--invalid-flag",
					"some-command",
				})

				Expect(err).To(Equal(application.NewInvalidFlagError(
					errors.New("flag provided but not defined: -invalid-flag"),
				)))
			})

			It("returns an error when the command is not passed in", func() {
				_, err := commandLineParser.Parse([]string{})
				Expect(err).To(MatchError(application.NewCommandNotProvidedError()))
			})

			It("returns an error when it cannot get working directory", func() {
				application.SetGetwd(func() (string, error) {
					return "", errors.New("failed to get working directory")
				})
				defer application.ResetGetwd()

				_, err := commandLineParser.Parse([]string{
					"some-command",
				})
				Expect(err).To(MatchError("failed to get working directory"))
			})
		})
	})
})
