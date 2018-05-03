package bosh

import (
	"io"
	"os/exec"
)

type CLI struct {
	stderr io.Writer
	path   string
}

func NewCLI(stderr io.Writer, path string) CLI {
	return CLI{
		stderr: stderr,
		path:   path,
	}
}

func (c CLI) Run(stdout io.Writer, workingDirectory string, args []string) error {
	command := exec.Command(c.path, args...)
	command.Dir = workingDirectory

	command.Stdout = stdout
	command.Stderr = c.stderr

	return command.Run()
}

func (c CLI) GetBOSHPath() string {
	return c.path
}
