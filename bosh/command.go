package bosh

import (
	"io"
	"os/exec"
)

type Cmd struct {
	stdout io.Writer
	stderr io.Writer
	debug  bool
}

func NewCmd(stdout, stderr io.Writer, debug bool) Cmd {
	return Cmd{
		stdout: stdout,
		stderr: stderr,
		debug:  debug,
	}
}

func (c Cmd) Run(workingDirectory string, args []string) error {
	command := exec.Command("bosh", args...)
	command.Dir = workingDirectory

	if c.debug {
		command.Stdout = c.stdout
		command.Stderr = c.stderr
	}

	return command.Run()
}
