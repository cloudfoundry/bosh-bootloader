package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

func main() {
	var options struct {
		Version bool `short:"v" long:"version" description:"Show version"`
	}

	_, err := flags.ParseArgs(&options, os.Args)
	if err != nil {
		os.Exit(1)
	}

	if options.Version {
		fmt.Println("bbl 0.0.1")
	}
}
