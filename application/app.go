package application

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
)

type CommandSet map[string]commands.Command

type App struct {
	commands CommandSet
	usage    func()
}

func New(commands CommandSet, usage func()) App {
	return App{
		commands: commands,
		usage:    usage,
	}
}

type config struct {
	Help    bool
	Version bool
	commands.GlobalFlags
}

func (a App) Run(args []string) error {
	globalFlags := flags.New("global")

	var cfg config
	globalFlags.Bool(&cfg.Help, "h", "help", false)
	globalFlags.Bool(&cfg.Version, "v", "version", false)
	globalFlags.String(&cfg.EndpointOverride, "endpoint-override", "")
	globalFlags.String(&cfg.AWSAccessKeyID, "aws-access-key-id", "")
	globalFlags.String(&cfg.AWSSecretAccessKey, "aws-secret-access-key", "")
	globalFlags.String(&cfg.AWSRegion, "aws-region", "")

	err := globalFlags.Parse(args)
	if err != nil {
		a.usage()
		return err
	}

	c := "[EMPTY]"
	if len(globalFlags.Args()) > 0 {
		c = globalFlags.Args()[0]
	}

	if cfg.Version {
		c = "version"
	}

	if cfg.Help {
		c = "help"
	}

	//^^^CONFIG^^^///vvvDOING SHITvvv//

	cmd, ok := a.commands[c]
	if !ok {
		a.usage()
		return fmt.Errorf("unknown command: %s", c)
	}

	err = cmd.Execute(cfg.GlobalFlags)
	if err != nil {
		return err
	}

	return nil
}
