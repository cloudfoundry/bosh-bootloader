package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

func main() {
	var options struct {
		Help    bool `short:"h" long:"help" description:"Show usage"`
		Version bool `short:"v" long:"version" description:"Show version"`
	}

	parser := flags.NewParser(&options, flags.PrintErrors|flags.PassDoubleDash)
	_, err := parser.ParseArgs(os.Args)
	if err != nil {
		os.Exit(1)
	}

	if options.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if options.Version {
		fmt.Println("bbl 0.0.1")
	}
}
