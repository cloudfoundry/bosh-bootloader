package terraform

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

func (c Cmd) Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error {
	command := exec.Command("terraform", args...)
	command.Dir = workingDirectory

	if debug {
		command.Stdout = stdout
		command.Stderr = c.stderr
	}

	return command.Run()
}
