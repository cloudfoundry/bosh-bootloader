package application_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type setNewKeyPairName struct{}

func (snkp setNewKeyPairName) Execute(subcommandFlags []string, state storage.State) (storage.State, error) {
	state.KeyPair = storage.KeyPair{
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

	var NewAppWithConfiguration = func(configuration application.Configuration) application.App {
		return application.New(application.CommandSet{
			"help":                 helpCmd,
			"version":              versionCmd,
			"some":                 someCmd,
			"error":                errorCmd,
			"set-new-keypair-name": setNewKeyPairName{},
		},
			configuration,
			stateStore,
			func() { helpCmd.Execute([]string{}, storage.State{}) })
	}

	BeforeEach(func() {
		helpCmd = &fakes.Command{}
		versionCmd = &fakes.Command{}
		errorCmd = &fakes.Command{}

		someCmd = &fakes.Command{}
		someCmd.ExecuteCall.PassState = true

		stateStore = &fakes.StateStore{}

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
						StateDir:         "some/state/dir",
						EndpointOverride: "some-endpoint-override",
					},
					State: storage.State{
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
						},
					},
				})

				Expect(app.Run()).To(Succeed())
				Expect(someCmd.ExecuteCall.CallCount).To(Equal(1))

				Expect(someCmd.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{
					"--first-subcommand-flag", "first-value",
					"--second-subcommand-flag", "second-value",
				}))
			})

			It("save the state", func() {
				app = NewAppWithConfiguration(application.Configuration{
					Command: "set-new-keypair-name",
					Global: application.GlobalConfiguration{
						StateDir: "some/state/dir",
					},
					State: storage.State{
						KeyPair: storage.KeyPair{
							Name:       "some-keypair-name",
							PrivateKey: "some-private-key",
						},
					},
				})

				Expect(app.Run()).To(Succeed())

				Expect(stateStore.SetCall.Receives.Dir).To(Equal("some/state/dir"))
				Expect(stateStore.SetCall.Receives.State.KeyPair).To(Equal(storage.KeyPair{
					Name:       "some-new-keypair-name",
					PrivateKey: "some-private-key",
				}))
			})

			It("saves the state even when the command fails", func() {
				errorCmd.ExecuteCall.Returns.Error = errors.New("error executing command")
				errorCmd.ExecuteCall.Returns.State = storage.State{
					EnvID: "some-env-time:stamp",
				}
				app = NewAppWithConfiguration(application.Configuration{
					Command: "error",
					Global: application.GlobalConfiguration{
						StateDir: "some/state/dir",
					},
					State: storage.State{
						KeyPair: storage.KeyPair{
							Name:       "some-keypair-name",
							PrivateKey: "some-private-key",
						},
					},
				})

				err := app.Run()
				Expect(err).To(MatchError("error executing command"))

				Expect(stateStore.SetCall.Receives.Dir).To(Equal("some/state/dir"))
				Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
					EnvID: "some-env-time:stamp",
				}))
			})
		})

		Context("error cases", func() {
			It("returns an error when the store can not be written to", func() {
				app = NewAppWithConfiguration(application.Configuration{
					Command: "some",
				})
				stateStore.SetCall.Returns.Error = errors.New("could not write to the store")
				err := app.Run()

				Expect(err).To(MatchError("could not write to the store"))
			})

			Context("when an unknown command is provided", func() {
				It("returns an error", func() {
					app = NewAppWithConfiguration(application.Configuration{
						Command: "some-unknown-command",
					})
					err := app.Run()
					Expect(err).To(MatchError("unknown command: some-unknown-command"))
					Expect(helpCmd.ExecuteCall.CallCount).To(Equal(1))
				})
			})

			Context("when the command fails to execute", func() {
				It("returns an error", func() {
					errorCmd.ExecuteCall.Returns.Error = errors.New("error executing command")
					app = NewAppWithConfiguration(application.Configuration{
						Command: "error",
					})
					err := app.Run()
					Expect(err).To(MatchError("error executing command"))
				})
			})

			Context("when both the command and writing the config fails", func() {
				It("doesn't mask the actual error, also reports the write failure", func() {

					errorCmd.ExecuteCall.Returns.Error = errors.New("error executing command")
					stateStore.SetCall.Returns.Error = errors.New("could not write configuration")
					app = NewAppWithConfiguration(application.Configuration{
						Command: "error",
						Global: application.GlobalConfiguration{
							StateDir: "some/state/dir",
						},
						State: storage.State{},
					})

					err := app.Run()
					Expect(err).To(MatchError(`"error" command failed with "error executing command", and the state failed to save with error "could not write configuration"`))
				})
			})
		})
	})
})
