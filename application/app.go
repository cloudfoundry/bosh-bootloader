package application

import (
	"fmt"
	"io"

	"github.com/jessevdk/go-flags"
)

type App struct{}

func New() App {
	return App{}
}

func (a App) Run(args []string, stdout io.Writer) error {
	var options struct {
		Help    bool `short:"h" long:"help" description:"Show usage"`
		Version bool `short:"v" long:"version" description:"Show version"`
	}

	parser := flags.NewParser(&options, flags.PassDoubleDash)
	_, err := parser.ParseArgs(args)
	if err != nil {
		return err
	}

	if options.Help {
		parser.WriteHelp(stdout)
		return nil
	}

	if options.Version {
		fmt.Fprintln(stdout, "bbl 0.0.1")
	}

	return nil
}
