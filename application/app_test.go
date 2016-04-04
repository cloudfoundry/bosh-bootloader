package application_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type setNewKeyPairName struct{}

func (snkp setNewKeyPairName) Execute(flags commands.GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	state.KeyPair = &storage.KeyPair{
		Name:       "some-new-keypair-name",
		PublicKey:  state.KeyPair.PublicKey,
		PrivateKey: state.KeyPair.PrivateKey,
	}

	return state, nil
}

var _ = Describe("App", func() {
	var (
		app        application.App
		helpCmd    *fakes.Command
		versionCmd *fakes.Command
		someCmd    *fakes.Command
		errorCmd   *fakes.Command
		stateStore *fakes.StateStore
	)

	BeforeEach(func() {
		helpCmd = &fakes.Command{}
		versionCmd = &fakes.Command{}
		errorCmd = &fakes.Command{}

		someCmd = &fakes.Command{}
		someCmd.ExecuteCall.PassState = true

		stateStore = &fakes.StateStore{}

		app = application.New(application.CommandSet{
			"help":                 helpCmd,
			"version":              versionCmd,
			"some":                 someCmd,
			"error":                errorCmd,
			"set-new-keypair-name": setNewKeyPairName{},
		},
			stateStore,
			func() { helpCmd.Execute(commands.GlobalFlags{}, []string{}, storage.State{}) })
	})

	Describe("Run", func() {
		Context("printing help", func() {
			It("prints out the usage when provided the --help flag", func() {
				Expect(app.Run([]string{"--help"})).To(Succeed())
				Expect(helpCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(helpCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{}))
			})

			It("prints out the usage when provided the -h flag", func() {
				Expect(app.Run([]string{"-h"})).To(Succeed())
				Expect(helpCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(helpCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{}))
			})

			It("prints out the usage when provided the help command", func() {
				Expect(app.Run([]string{"help"})).To(Succeed())
				Expect(helpCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(helpCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{}))
			})
		})

		Context("printing version", func() {
			It("prints out the current version when provided the -v flag", func() {
				Expect(app.Run([]string{"-v"})).To(Succeed())
				Expect(versionCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(versionCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{}))
			})

			It("prints out the current version when provided the --version flag", func() {
				Expect(app.Run([]string{"--version"})).To(Succeed())
				Expect(versionCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(versionCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{}))
			})

			It("prints out the current version when provided the version command", func() {
				Expect(app.Run([]string{"version"})).To(Succeed())
				Expect(versionCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(versionCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{}))
			})
		})

		Context("executing arbitrary commands", func() {
			It("executes the correct command", func() {
				Expect(app.Run([]string{"some"})).To(Succeed())
				Expect(someCmd.ExecuteCall.CallCount).To(Equal(1))
				Expect(someCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{}))
			})

			It("save state when the command has modified the state", func() {
				stateStore.GetCall.Returns.State = storage.State{
					KeyPair: &storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-private-key",
					},
				}

				Expect(app.Run([]string{
					"--endpoint-override", "some-endpoint-override",
					"--aws-access-key-id", "some-aws-access-key-id",
					"--aws-secret-access-key", "some-aws-secret-access-key",
					"--aws-region", "some-aws-region",
					"--state-dir", "/some/state/dir",
					"set-new-keypair-name",
				})).To(Succeed())

				Expect(stateStore.SetCall.Receives.State.KeyPair).To(Equal(&storage.KeyPair{
					Name:       "some-new-keypair-name",
					PrivateKey: "some-private-key",
				}))
			})

			Context("when subcommand flags are provided", func() {
				It("passes the flags to the subcommand", func() {
					Expect(app.Run([]string{
						"some",
						"--first-subcommand-flag", "first-value",
						"--second-subcommand-flag", "second-value",
					})).To(Succeed())

					Expect(someCmd.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{
						"--first-subcommand-flag", "first-value",
						"--second-subcommand-flag", "second-value",
					}))
				})
			})

			Context("when global flags are provided", func() {
				It("stores the flags in the state store", func() {
					stateStore.GetCall.Returns.State = storage.State{
						KeyPair: &storage.KeyPair{
							Name: "some-keypair-name",
						},
					}

					Expect(app.Run([]string{
						"--endpoint-override", "some-endpoint-override",
						"--aws-access-key-id", "some-aws-access-key-id",
						"--aws-secret-access-key", "some-aws-secret-access-key",
						"--aws-region", "some-aws-region",
						"--state-dir", "/some/state/dir",
						"some",
					})).To(Succeed())

					Expect(stateStore.GetCall.Receives.Dir).To(Equal("/some/state/dir"))

					Expect(stateStore.SetCall.Receives.Dir).To(Equal("/some/state/dir"))
					Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
						AWS: storage.AWS{
							Region:          "some-aws-region",
							SecretAccessKey: "some-aws-secret-access-key",
							AccessKeyID:     "some-aws-access-key-id",
						},
						KeyPair: &storage.KeyPair{
							Name: "some-keypair-name",
						},
					}))
				})

				Context("when the state has not changed", func() {
					It("does not store the state again", func() {
						stateStore.GetCall.Returns.State = storage.State{
							KeyPair: &storage.KeyPair{
								Name: "some-new-keypair-name",
							},
						}

						Expect(app.Run([]string{
							"--endpoint-override", "some-endpoint-override",
							"--state-dir", "/some/state/dir",
							"set-new-keypair-name",
						})).To(Succeed())

						Expect(stateStore.GetCall.Receives.Dir).To(Equal("/some/state/dir"))
						Expect(stateStore.SetCall.CallCount).To(Equal(0))
					})
				})

				It("executes the command with those flags", func() {
					Expect(app.Run([]string{
						"--endpoint-override", "some-endpoint-override",
						"--aws-access-key-id", "some-aws-access-key-id",
						"--aws-secret-access-key", "some-aws-secret-access-key",
						"--aws-region", "some-aws-region",
						"--state-dir", "/some/state/dir",
						"some",
					})).To(Succeed())
					Expect(someCmd.ExecuteCall.CallCount).To(Equal(1))
					Expect(someCmd.ExecuteCall.Receives.GlobalFlags).To(Equal(commands.GlobalFlags{
						EndpointOverride:   "some-endpoint-override",
						AWSAccessKeyID:     "some-aws-access-key-id",
						AWSSecretAccessKey: "some-aws-secret-access-key",
						AWSRegion:          "some-aws-region",
						StateDir:           "/some/state/dir",
					}))
				})
			})
		})

		Context("error cases", func() {
			It("returns an error when the store can not be read from", func() {
				stateStore.GetCall.Returns.Error = errors.New("could not read from store")
				err := app.Run([]string{
					"--state-dir", "/some/state/dir",
				})

				Expect(err).To(MatchError("could not read from store"))
			})

			It("returns an error when the store can not be written to", func() {
				stateStore.SetCall.Returns.Error = errors.New("could not write to the store")
				err := app.Run([]string{
					"--aws-region", "some-aws-region",
					"--state-dir", "/some/state/dir",
					"some",
				})

				Expect(err).To(MatchError("could not write to the store"))
			})

			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					err := app.Run([]string{"--some-unknown-flag"})
					Expect(err).To(MatchError("flag provided but not defined: -some-unknown-flag"))
					Expect(helpCmd.ExecuteCall.CallCount).To(Equal(1))
				})
			})

			Context("when an unknown command is provided", func() {
				It("returns an error", func() {
					err := app.Run([]string{"unknown-command"})
					Expect(err).To(MatchError("unknown command: unknown-command"))
					Expect(helpCmd.ExecuteCall.CallCount).To(Equal(1))
				})
			})

			Context("when nothing is provided", func() {
				It("returns an error", func() {
					err := app.Run([]string{})
					Expect(err).To(MatchError("unknown command: [EMPTY]"))
					Expect(helpCmd.ExecuteCall.CallCount).To(Equal(1))
				})
			})

			Context("When the command fails to execute", func() {
				It("returns an error", func() {
					errorCmd.ExecuteCall.Returns.Error = errors.New("error executing command")
					err := app.Run([]string{"error"})
					Expect(err).To(MatchError("error executing command"))
				})
			})
		})
	})
})
