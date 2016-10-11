package commands

import (
	"fmt"
	"io"
	"runtime"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	VersionCommand = "version"
	BBLDevVersion  = "dev"
)

type Version struct {
	stdout  io.Writer
	version string
}

func NewVersion(version string, stdout io.Writer) Version {
	if version == "" {
		version = BBLDevVersion
	}
	return Version{
		stdout:  stdout,
		version: fmt.Sprintf("%s (%s/%s)", version, runtime.GOOS, runtime.GOARCH),
	}
}

func (v Version) Execute(subcommandFlags []string, state storage.State) error {
	fmt.Fprintln(v.stdout, fmt.Sprintf("bbl %s", v.version))
	return nil
}
