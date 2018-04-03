package bosh

import (
	"io"
	"os/exec"
)

type Cmd struct {
	stderr io.Writer
	path   string
}

func NewCmd(stderr io.Writer, path string) Cmd {
	return Cmd{
		stderr: stderr,
		path:   path,
	}
}

func (c Cmd) Run(stdout io.Writer, workingDirectory string, args []string) error {
	command := exec.Command(c.path, args...)
	command.Dir = workingDirectory

	command.Stdout = stdout
	command.Stderr = c.stderr

	return command.Run()
}

func (c Cmd) GetBOSHPath() string {
	return c.path
}
