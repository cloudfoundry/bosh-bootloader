package commands

import (
	"fmt"
	"io"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const VERSION = "bbl 0.0.1"

type Version struct {
	stdout io.Writer
}

func NewVersion(stdout io.Writer) Version {
	return Version{stdout}
}

func (v Version) Execute(globalFlags GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	fmt.Fprintln(v.stdout, VERSION)
	return state, nil
}
