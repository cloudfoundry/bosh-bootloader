package commands

import (
	"fmt"
	"io"
)

const VERSION = "bbl 0.0.1"

type Version struct {
	stdout io.Writer
}

func NewVersion(stdout io.Writer) Version {
	return Version{stdout}
}

func (v Version) Execute(globalFlags GlobalFlags) error {
	fmt.Fprint(v.stdout, VERSION)
	return nil
}
