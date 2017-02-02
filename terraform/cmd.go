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

func (cmd Cmd) Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error {
	runCommand := exec.Command("terraform", args...)
	runCommand.Dir = workingDirectory

	if debug {
		runCommand.Stdout = stdout
		runCommand.Stderr = cmd.stderr
	}

	return runCommand.Run()
}
