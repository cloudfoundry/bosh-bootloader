package terraform

import (
	"io"
	"os/exec"
)

type Cmd struct {
	stdout io.Writer
	stderr io.Writer
}

func NewCmd(stdout, stderr io.Writer) Cmd {
	return Cmd{
		stdout: stdout,
		stderr: stderr,
	}
}

func (cmd Cmd) Run(args []string) error {
	runCommand := exec.Command("terraform", args...)
	runCommand.Stdout = cmd.stdout
	runCommand.Stderr = cmd.stderr

	return runCommand.Run()
}
