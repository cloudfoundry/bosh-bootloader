package application_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandLineParser", func() {
	var (
		commandLineParser application.CommandLineParser
		usageCallCount    int
	)

	BeforeEach(func() {
		usageCallCount = 0
		usageFunc := func() {
			usageCallCount++
		}
		commandLineParser = application.NewCommandLineParser(usageFunc)
	})

	Describe("Parse", func() {
		It("returns a command line configuration with correct global flags based on arguments passed in", func() {
			args := []string{
				"--endpoint-override=some-endpoint-override",
				"--state-dir", "some/state/dir",
				"some-command",
				"--subcommand-flag", "some-value",
			}
			commandLineConfiguration, err := commandLineParser.Parse(args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandLineConfiguration.EndpointOverride).To(Equal("some-endpoint-override"))
			Expect(commandLineConfiguration.StateDir).To(Equal("some/state/dir"))
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
			It("returns an error and prints usage when an invalid flag is provided", func() {
				_, err := commandLineParser.Parse([]string{
					"--invalid-flag",
					"some-command",
				})

				Expect(err).To(Equal(errors.New("flag provided but not defined: -invalid-flag")))
				Expect(usageCallCount).To(Equal(1))
			})

			It("returns an error and prints usage when command is not provided", func() {
				_, err := commandLineParser.Parse([]string{})

				Expect(err).To(Equal(errors.New("unknown command: [EMPTY]")))
				Expect(usageCallCount).To(Equal(1))
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
