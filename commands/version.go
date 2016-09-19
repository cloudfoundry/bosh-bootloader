package commands

import (
	"fmt"
	"io"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const (
	VersionCommand = "version"
	BBLVersion     = "bbl 0.0.1"
)

type Version struct {
	stdout io.Writer
}

func NewVersion(stdout io.Writer) Version {
	return Version{stdout}
}

func (v Version) Execute(subcommandFlags []string, state storage.State) error {
	fmt.Fprintln(v.stdout, BBLVersion)
	return nil
}
