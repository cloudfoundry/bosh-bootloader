package bosh

import (
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

func (c Cmd) Run(stdout io.Writer, workingDirectory string, args []string) error {
	command := exec.Command("bosh", args...)
	command.Dir = workingDirectory

	command.Stdout = stdout
	command.Stderr = c.stderr

	return command.Run()
}
