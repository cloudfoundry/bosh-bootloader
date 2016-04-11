package application

import (
	"fmt"
	"os"
	"reflect"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

var getwd func() (string, error) = os.Getwd

type CommandSet map[string]commands.Command

type store interface {
	Get(dir string) (storage.State, error)
	Set(dir string, state storage.State) error
}

type App struct {
	commands CommandSet
	store    store
	usage    func()
}

func New(commands CommandSet, store store, usage func()) App {
	return App{
		commands: commands,
		store:    store,
		usage:    usage,
	}
}

type config struct {
	Command         string
	SubcommandFlags []string
	Help            bool
	Version         bool
	commands.GlobalFlags
}

func (a App) Run(args []string) error {
	cfg, err := a.configure(args)
	if err != nil {
		return err
	}

	state, err := a.store.Get(cfg.GlobalFlags.StateDir)
	if err != nil {
		return err
	}

	newState, err := a.execute(cfg, cfg.SubcommandFlags, a.applyGlobalConfig(cfg.GlobalFlags, state))
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(newState, state) {
		err = a.store.Set(cfg.GlobalFlags.StateDir, newState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a App) configure(args []string) (config, error) {
	globalFlags := flags.New("global")

	cfg := config{
		Command: "[EMPTY]",
	}
	globalFlags.Bool(&cfg.Help, "h", "help", false)
	globalFlags.Bool(&cfg.Version, "v", "version", false)
	globalFlags.String(&cfg.EndpointOverride, "endpoint-override", "")
	globalFlags.String(&cfg.AWSAccessKeyID, "aws-access-key-id", "")
	globalFlags.String(&cfg.AWSSecretAccessKey, "aws-secret-access-key", "")
	globalFlags.String(&cfg.AWSRegion, "aws-region", "")
	globalFlags.String(&cfg.StateDir, "state-dir", "")

	err := globalFlags.Parse(args)
	if err != nil {
		a.usage()
		return cfg, err
	}

	if len(globalFlags.Args()) > 0 {
		cfg.Command = globalFlags.Args()[0]
		cfg.SubcommandFlags = globalFlags.Args()[1:]
	}

	if cfg.Version {
		cfg.Command = "version"
	}

	if cfg.Help {
		cfg.Command = "help"
	}

	if cfg.GlobalFlags.StateDir == "" {
		wd, err := getwd()
		if err != nil {
			return cfg, err
		}

		cfg.GlobalFlags.StateDir = wd
	}

	return cfg, nil
}

func (a App) applyGlobalConfig(globals commands.GlobalFlags, state storage.State) storage.State {
	if globals.AWSAccessKeyID != "" {
		state.AWS.AccessKeyID = globals.AWSAccessKeyID
	}

	if globals.AWSSecretAccessKey != "" {
		state.AWS.SecretAccessKey = globals.AWSSecretAccessKey
	}

	if globals.AWSRegion != "" {
		state.AWS.Region = globals.AWSRegion
	}

	return state
}

func (a App) execute(cfg config, subcommandFlags []string, state storage.State) (storage.State, error) {
	cmd, ok := a.commands[cfg.Command]
	if !ok {
		a.usage()
		return state, fmt.Errorf("unknown command: %s", cfg.Command)
	}

	state, err := cmd.Execute(cfg.GlobalFlags, subcommandFlags, state)
	if err != nil {
		return state, err
	}

	return state, nil
}
