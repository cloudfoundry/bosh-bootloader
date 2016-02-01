package main

import (
	"fmt"
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
)

func main() {
	app := application.New()

	err := app.Run(os.Args, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}
