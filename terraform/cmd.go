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

func (cmd Cmd) Run(workingDirectory string, args []string) error {
	runCommand := exec.Command("terraform", args...)
	runCommand.Dir = workingDirectory
	runCommand.Stdout = cmd.stdout
	runCommand.Stderr = cmd.stderr

	return runCommand.Run()
}
