package application_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandLineParser", func() {
	var (
		commandLineParser application.CommandLineParser
		usageCallCount    int
		fakeEnvGetter     *fakes.EnvGetter
	)

	BeforeEach(func() {
		usageCallCount = 0
		usageFunc := func() {
			usageCallCount++
		}
		commandSet := application.CommandSet{
			commands.UpCommand:      nil,
			commands.VersionCommand: nil,
			commands.HelpCommand:    nil,
		}
		fakeEnvGetter = &fakes.EnvGetter{}

		commandLineParser = application.NewCommandLineParser(usageFunc, commandSet, fakeEnvGetter)
	})

	Describe("Parse", func() {
		It("returns a command line configuration with correct global flags based on arguments passed in", func() {
			args := []string{
				"--state-dir", "some/state/dir",
				"--debug",
				"up",
				"--subcommand-flag", "some-value",
			}
			commandLineConfiguration, err := commandLineParser.Parse(args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandLineConfiguration.StateDir).To(Equal("some/state/dir"))
			Expect(commandLineConfiguration.Debug).To(BeTrue())
		})

		It("returns a command line configuration with correct command with subcommand flags based on arguments passed in", func() {
			args := []string{
				"up",
				"--subcommand-flag", "some-value",
			}
			commandLineConfiguration, err := commandLineParser.Parse(args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandLineConfiguration.Command).To(Equal("up"))
			Expect(commandLineConfiguration.SubcommandFlags).To(Equal([]string{"--subcommand-flag", "some-value"}))
		})

		Context("when the BBL_DEBUG environment variable is provided", func() {
			BeforeEach(func() {
				fakeEnvGetter.Values = map[string]string{
					"BBL_DEBUG": "true",
				}
			})

			It("returns a command line configuration with the debug configuration set to true", func() {
				args := []string{
					"--state-dir", "some/state/dir",
					"up",
					"--subcommand-flag", "some-value",
				}
				commandLineConfiguration, err := commandLineParser.Parse(args)
				Expect(err).NotTo(HaveOccurred())

				Expect(commandLineConfiguration.StateDir).To(Equal("some/state/dir"))
				Expect(commandLineConfiguration.Debug).To(BeTrue())
			})
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
					"up",
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

			Entry("returns the version command provided version", "version", "version"),
			Entry("returns the version command provided --version", "--version", "version"),
		)

		It("runs help without error if more arguments are provided to help", func() {
			commandLineConfiguration, err := commandLineParser.Parse([]string{
				"--help",
				"up",
				"--aws-stuff",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(commandLineConfiguration.Command).To(Equal("help"))
			Expect(commandLineConfiguration.SubcommandFlags).To(Equal([]string{"up", "--aws-stuff"}))
		})

		It("runs help without error", func() {
			commandLineConfiguration, err := commandLineParser.Parse([]string{})

			Expect(err).NotTo(HaveOccurred())
			Expect(commandLineConfiguration.Command).To(Equal("help"))
			Expect(commandLineConfiguration.SubcommandFlags).To(BeEmpty())
		})

		Context("failure cases", func() {
			It("returns an error and prints usage when an invalid flag is provided", func() {
				_, err := commandLineParser.Parse([]string{
					"--invalid-flag",
					"up",
				})

				Expect(err).To(Equal(errors.New("flag provided but not defined: -invalid-flag")))
				Expect(usageCallCount).To(Equal(1))
			})

			It("returns an error and prints usage when an invalid flag is provided to help", func() {
				_, err := commandLineParser.Parse([]string{
					"--help",
					"badcmd",
				})

				Expect(err).To(Equal(errors.New("Unrecognized command 'badcmd'")))
				Expect(usageCallCount).To(Equal(1))
			})

			It("returns an error when it cannot get working directory", func() {
				application.SetGetwd(func() (string, error) {
					return "", errors.New("failed to get working directory")
				})
				defer application.ResetGetwd()

				_, err := commandLineParser.Parse([]string{
					"up",
				})
				Expect(err).To(MatchError("failed to get working directory"))
			})

			It("validates the command before it validates the global arguments", func() {
				_, err := commandLineParser.Parse([]string{
					"--badflag", "x", "help", "delete-lbs", "--other-flag",
				})

				Expect(err).To(Equal(errors.New("Unrecognized command 'x'")))
				Expect(usageCallCount).To(Equal(1))
			})
		})
	})
})
