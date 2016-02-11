package application

import (
	"io"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
)

type App struct{}

func New() App {
	return App{}
}

func (a App) Run(args []string, stdout io.Writer) error {
	var parser *flags.Parser

	bblCommand := commands.NewBBLCommand(stdout)

	bblCommand.Help = func() {
		parser.WriteHelp(stdout)
		os.Exit(0)
	}

	parser = flags.NewParser(bblCommand, flags.PassDoubleDash)
	_, err := parser.ParseArgs(args)
	if err != nil {
		return err
	}

	return nil
}
