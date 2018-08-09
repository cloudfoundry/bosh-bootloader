package application_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("App", func() {
	var (
		app        application.App
		helpCmd    *fakes.Command
		versionCmd *fakes.Command
		someCmd    *fakes.Command
		errorCmd   *fakes.Command
		usage      *fakes.Usage
	)

	var NewAppWithConfiguration = func(configuration application.Configuration) application.App {
		return application.New(application.CommandSet{
			"help":      helpCmd,
			"version":   versionCmd,
			"--version": versionCmd,
			"some":      someCmd,
			"error":     errorCmd,
		},
			configuration,
			usage,
		)
	}

	BeforeEach(func() {
		helpCmd = &fakes.Command{}
		versionCmd = &fakes.Command{}
		errorCmd = &fakes.Command{}

		someCmd = &fakes.Command{}
		someCmd.ExecuteCall.PassState = true

		usage = &fakes.Usage{}

		app = NewAppWithConfiguration(application.Configuration{})
	})

	Describe("Run", func() {
		Context("executing commands", func() {
			It("executes the command with flags", func() {
				app = NewAppWithConfiguration(application.Configuration{
					Command: "some",
					SubcommandFlags: []string{
						"--first-subcommand-flag", "first-value",
						"--second-subcommand-flag", "second-value",
					},
					Global: application.GlobalConfiguration{
						StateDir: "some/state/dir",
					},
					State: storage.State{},
				})

				Expect(app.Run()).To(Succeed())

				Expect(someCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(someCmd.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{
					"--first-subcommand-flag", "first-value",
					"--second-subcommand-flag", "second-value",
				}))
			})
		})

		Context("when name is passed as a global flag", func() {
			DescribeTable("propagates name to subcommand flags", func(command string) {
				commandFake := &fakes.Command{}

				app = application.New(
					application.CommandSet{
						command: commandFake,
					},
					application.Configuration{
						Command: command,
						SubcommandFlags: []string{
							"--first-subcommand-flag", "first-value",
							"--second-subcommand-flag", "second-value",
						},
						Global: application.GlobalConfiguration{
							StateDir: "some/state/dir",
							Name:     "some-env-name",
						},
						State: storage.State{},
					},
					usage)

				Expect(app.Run()).To(Succeed())

				Expect(commandFake.ExecuteCall.CallCount).To(Equal(1))
				Expect(commandFake.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{
					"--first-subcommand-flag", "first-value",
					"--second-subcommand-flag", "second-value",
					"--name", "some-env-name",
				}))
			},
				Entry("when command is plan", "plan"),
				Entry("when command is up", "up"),
			)
		})

		Context("when subcommand flags contains help", func() {
			DescribeTable("prints command specific usage when help subcommand flag is provided", func(helpFlag string) {
				someCmd.UsageCall.Returns.Usage = "some usage message"

				app = NewAppWithConfiguration(application.Configuration{
					Command:         "some",
					SubcommandFlags: []string{helpFlag},
					ShowCommandHelp: true,
				})

				Expect(app.Run()).To(Succeed())
				Expect(someCmd.UsageCall.CallCount).To(Equal(1))
				Expect(usage.PrintCommandUsageCall.CallCount).To(Equal(1))
				Expect(usage.PrintCommandUsageCall.Receives.Message).To(Equal("some usage message"))
				Expect(usage.PrintCommandUsageCall.Receives.Command).To(Equal("some"))
				Expect(someCmd.ExecuteCall.CallCount).To(Equal(0))
			},
				Entry("when --help is provided", "--help"),
				Entry("when -h is provided", "-h"),
			)
		})

		Context("when help is called with a command", func() {
			It("prints the command specific help", func() {
				someCmd.UsageCall.Returns.Usage = "some usage message"

				app = NewAppWithConfiguration(application.Configuration{
					Command:         "help",
					SubcommandFlags: []string{"some"},
				})

				Expect(app.Run()).To(Succeed())
				Expect(someCmd.UsageCall.CallCount).To(Equal(1))
				Expect(usage.PrintCommandUsageCall.CallCount).To(Equal(1))
				Expect(usage.PrintCommandUsageCall.Receives.Message).To(Equal("some usage message"))
				Expect(usage.PrintCommandUsageCall.Receives.Command).To(Equal("some"))
				Expect(someCmd.ExecuteCall.CallCount).To(Equal(0))
			})

			Context("failure cases", func() {
				Context("when a invalid subcommand is passed", func() {
					BeforeEach(func() {
						app = NewAppWithConfiguration(application.Configuration{
							Command:         "help",
							SubcommandFlags: []string{"invalid-command"},
						})
					})

					It("prints the usage", func() {
						err := app.Run()
						Expect(err).To(MatchError("unknown command: invalid-command"))
						Expect(someCmd.ExecuteCall.CallCount).To(Equal(0))
						Expect(usage.PrintCall.CallCount).To(Equal(1))
					})
				})
			})
		})

		Context("when --version is the command", func() {
			It("executes the command", func() {
				app = NewAppWithConfiguration(application.Configuration{
					Command: "--version",
					SubcommandFlags: []string{
						"--first-subcommand-flag", "first-value",
						"--second-subcommand-flag", "second-value",
					},
				})

				Expect(app.Run()).To(Succeed())

				Expect(versionCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(versionCmd.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{}))
			})
		})

		Context("when subcommand flags contains version", func() {
			DescribeTable("prints version when version subcommand flag is provided", func(versionFlag string) {
				app = NewAppWithConfiguration(application.Configuration{
					Command:         "some",
					SubcommandFlags: []string{versionFlag},
				})

				Expect(app.Run()).To(Succeed())
				Expect(someCmd.ExecuteCall.CallCount).To(Equal(0))
				Expect(versionCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(versionCmd.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{}))
				Expect(versionCmd.ExecuteCall.Receives.State).To(Equal(storage.State{}))
			},
				Entry("when --version is provided", "--version"),
				Entry("when -v is provided", "-v"),
			)

			Context("error cases", func() {
				Context("when version command is not part of the command set", func() {
					BeforeEach(func() {
						app = application.New(application.CommandSet{
							"some": someCmd,
						}, application.Configuration{
							Command:         "some",
							SubcommandFlags: []string{"-v"},
						}, usage)
					})

					It("returns an error", func() {
						Expect(app.Run()).To(MatchError("unknown command: version"))
					})
				})
			})
		})

		Context("error cases", func() {
			Context("when a fast fail occurs", func() {
				BeforeEach(func() {
					someCmd.CheckFastFailsCall.Returns.Error = errors.New("fast failed command")
				})

				It("returns an error and does not execute the command", func() {
					app = NewAppWithConfiguration(application.Configuration{
						Command: "some",
					})
					err := app.Run()
					Expect(someCmd.CheckFastFailsCall.CallCount).To(Equal(1))
					Expect(err).To(MatchError("fast failed command"))
					Expect(someCmd.ExecuteCall.CallCount).To(Equal(0))
				})
			})

			Context("when a fast fail occurs, but indicates that its a success", func() {
				BeforeEach(func() {
					someCmd.CheckFastFailsCall.Returns.Error = commands.ExitSuccessfully{}
				})

				It("returns nil and doesn't execute the command", func() {
					app = NewAppWithConfiguration(application.Configuration{
						Command: "some",
					})
					err := app.Run()
					Expect(someCmd.CheckFastFailsCall.CallCount).To(Equal(1))
					Expect(err).NotTo(HaveOccurred())
					Expect(someCmd.ExecuteCall.CallCount).To(Equal(0))
				})
			})

			Context("when an unknown command is provided", func() {
				It("prints usage and returns an error", func() {
					app = NewAppWithConfiguration(application.Configuration{
						Command: "some-unknown-command",
					})
					err := app.Run()
					Expect(err).To(MatchError("unknown command: some-unknown-command"))
					Expect(usage.PrintCall.CallCount).To(Equal(1))
				})
			})

			Context("when the command fails to execute", func() {
				It("returns an error", func() {
					errorCmd.ExecuteCall.Returns.Error = errors.New("error executing command")
					app = NewAppWithConfiguration(application.Configuration{
						Command: "error",
						Global: application.GlobalConfiguration{
							Debug: true,
						},
					})
					Expect(app.Run()).To(MatchError("error executing command"))
				})
			})
		})
	})
})
