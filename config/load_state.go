package config

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/jessevdk/go-flags"
)

type logger interface {
	Println(string)
}

type StateBootstrap interface {
	GetState(string) (storage.State, error)
}

type migrator interface {
	Migrate(storage.State) (storage.State, error)
}

type merger interface {
	MergeGlobalFlagsToState(globalflags GlobalFlags, state storage.State) (storage.State, error)
}

type downloader interface {
	DownloadAndPrepareState(globalflags GlobalFlags) error
}

type fs interface {
	fileio.Stater
	fileio.TempFiler
	fileio.FileReader
	fileio.FileWriter
}

func NewConfig(bootstrap StateBootstrap, migrator migrator, merger merger, downloader downloader, logger logger, fs fs) Config {
	return Config{
		stateBootstrap: bootstrap,
		migrator:       migrator,
		merger:         merger,
		downloader:     downloader,
		logger:         logger,
		fs:             fs,
	}
}

type Config struct {
	stateBootstrap StateBootstrap
	migrator       migrator
	merger         merger
	downloader     downloader
	logger         logger
	fs             fs
}

func ParseArgs(args []string) (GlobalFlags, []string, error) {
	var globals GlobalFlags
	parser := flags.NewParser(&globals, flags.IgnoreUnknown)

	remainingArgs, err := parser.ParseArgs(args[1:])
	if err != nil {
		return GlobalFlags{}, remainingArgs, err
	}

	if globals.StateBucket != "" && globals.StateDir == "" {
		tempDir, err := os.MkdirTemp("", "bbl-state")
		if err != nil {
			return GlobalFlags{}, remainingArgs, err // not tested
		}
		globals.StateDir = tempDir

	} else if !filepath.IsAbs(globals.StateDir) {
		workingDir, err := os.Getwd()
		if err != nil {
			return GlobalFlags{}, remainingArgs, err // not tested
		}
		globals.StateDir = filepath.Join(workingDir, globals.StateDir)
	}

	return globals, remainingArgs, nil
}

func (c Config) Bootstrap(globalFlags GlobalFlags, remainingArgs []string, argsLen int) (application.Configuration, error) {
	if argsLen == 1 {
		return application.Configuration{
			Command: "help",
		}, nil
	}

	var command string
	if len(remainingArgs) > 0 {
		command = remainingArgs[0]
	}

	if globalFlags.Version || command == "version" {
		command = "version"
		return application.Configuration{
			ShowCommandHelp: globalFlags.Help,
			Command:         command,
		}, nil
	}

	if len(remainingArgs) == 0 {
		return application.Configuration{
			Command: "help",
		}, nil
	}

	if len(remainingArgs) == 1 && command == "help" {
		return application.Configuration{
			Command: command,
		}, nil
	}

	if command == "help" {
		return application.Configuration{
			ShowCommandHelp: true,
			Command:         remainingArgs[1],
		}, nil
	}

	if globalFlags.Help {
		return application.Configuration{
			ShowCommandHelp: true,
			Command:         command,
		}, nil
	}

	if !modifiesState(command) && globalFlags.StateBucket != "" {
		err := c.downloader.DownloadAndPrepareState(globalFlags)
		if err != nil {
			return application.Configuration{}, err
		}
	}

	state, err := c.stateBootstrap.GetState(globalFlags.StateDir)
	if err != nil {
		return application.Configuration{}, err
	}

	state, err = c.migrator.Migrate(state)
	if err != nil {
		return application.Configuration{}, err
	}

	state, err = c.merger.MergeGlobalFlagsToState(globalFlags, state)
	if err != nil {
		return application.Configuration{}, err
	}

	if modifiesState(command) {
		err = ValidateIAAS(state)
		if err != nil {
			return application.Configuration{}, err
		}
	}

	return application.Configuration{
		Global: application.GlobalConfiguration{
			Debug:    globalFlags.Debug,
			StateDir: globalFlags.StateDir,
			Name:     globalFlags.EnvID,
		},
		State:                state,
		Command:              command,
		SubcommandFlags:      remainingArgs[1:],
		ShowCommandHelp:      false,
		CommandModifiesState: modifiesState(command),
	}, nil
}

func modifiesState(command string) bool {
	_, ok := map[string]struct{}{ // membership in this is untested
		"up":                {},
		"down":              {},
		"plan":              {},
		"destroy":           {},
		"leftovers":         {},
		"cleanup-leftovers": {},
		"rotate":            {},
	}[command]
	return ok
}
