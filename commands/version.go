package commands

import (
	"fmt"
	"io"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	VersionCommand = "version"
	BBLDevVersion  = "dev"
)

type Version struct {
	version string
	stdout  io.Writer
}

func NewVersion(version string, stdout io.Writer) Version {
	return Version{
		version: version,
		stdout:  stdout,
	}
}

func (v Version) Execute(subcommandFlags []string, state storage.State) error {
	version := v.version
	if version == "" {
		version = BBLDevVersion
	}

	fmt.Fprintln(v.stdout, fmt.Sprintf("bbl %s", version))
	return nil
}
