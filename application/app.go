package application

import (
	"fmt"
	"reflect"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"
)

type CommandSet map[string]commands.Command

type store interface {
	Get(dir string) (state.State, error)
	Set(dir string, s state.State) error
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
	Command string
	Help    bool
	Version bool
	commands.GlobalFlags
}

func (a App) Run(args []string) error {
	cfg, err := a.configure(args)
	if err != nil {
		return err
	}

	s, err := a.store.Get(cfg.GlobalFlags.StateDir)
	if err != nil {
		return err
	}

	newState, err := a.execute(cfg, a.applyGlobalConfig(cfg.GlobalFlags, s))
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(newState, s) {
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
	}

	if cfg.Version {
		cfg.Command = "version"
	}

	if cfg.Help {
		cfg.Command = "help"
	}

	return cfg, nil
}

func (a App) applyGlobalConfig(globals commands.GlobalFlags, s state.State) state.State {
	if globals.AWSAccessKeyID != "" {
		s.AWS.AccessKeyID = globals.AWSAccessKeyID
	}

	if globals.AWSSecretAccessKey != "" {
		s.AWS.SecretAccessKey = globals.AWSSecretAccessKey
	}

	if globals.AWSRegion != "" {
		s.AWS.Region = globals.AWSRegion
	}

	return s
}

func (a App) execute(cfg config, s state.State) (state.State, error) {
	cmd, ok := a.commands[cfg.Command]
	if !ok {
		a.usage()
		return s, fmt.Errorf("unknown command: %s", cfg.Command)
	}

	s, err := cmd.Execute(cfg.GlobalFlags, s)
	if err != nil {
		return s, err
	}

	return s, nil
}
