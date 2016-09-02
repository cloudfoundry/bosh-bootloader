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

func (snkp setNewKeyPairName) Execute(subcommandFlags []string, state storage.State) error {
	state.KeyPair = storage.KeyPair{
		Name:       "some-new-keypair-name",
		PublicKey:  state.KeyPair.PublicKey,
		PrivateKey: state.KeyPair.PrivateKey,
	}

	return nil
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
		})

		Context("error cases", func() {
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
		})
	})
})
