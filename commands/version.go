package commands

import (
	"fmt"
	"runtime"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Version struct {
	logger  logger
	version string
}

func NewVersion(version string, logger logger) Version {
	return Version{
		logger:  logger,
		version: fmt.Sprintf("%s (%s/%s)", version, runtime.GOOS, runtime.GOARCH),
	}
}

func (v Version) Execute(subcommandFlags []string, state storage.State) error {
	v.logger.Printf("bbl %s\n", v.version)
	return nil
}

func (v Version) CheckFastFails(subcommandFlags []string, state storage.State) error {
	return nil
}
