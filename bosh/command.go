package bosh

import (
	"fmt"
	"io"
	"os/exec"
)

type Cmd struct {
	stderr io.Writer
}

func NewCmd(stderr io.Writer) Cmd {
	return Cmd{
		stderr: stderr,
	}
}

func (c Cmd) GetBOSHPath() (string, error) {
	var boshPath = "bosh"

	path, err := exec.LookPath("bosh2")
	if err != nil {
		if err.(*exec.Error).Err != exec.ErrNotFound {
			return "", fmt.Errorf("failed when searching for BOSH: %s", err) // not tested
		}
	}

	if path != "" {
		boshPath = path
	}

	return boshPath, nil
}

func (c Cmd) Run(stdout io.Writer, workingDirectory string, args []string) error {
	boshPath, err := c.GetBOSHPath()
	if err != nil {
		return err
	}

	command := exec.Command(boshPath, args...)
	command.Dir = workingDirectory

	command.Stdout = stdout
	command.Stderr = c.stderr

	return command.Run()
}
