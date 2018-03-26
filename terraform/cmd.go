package terraform

import (
	"io"
	"os"
	"os/exec"
)

type Cmd struct {
	stderr       io.Writer
	outputBuffer io.Writer
	tfDataDir    string
}

func NewCmd(stderr, outputBuffer io.Writer, tfDataDir string) Cmd {
	return Cmd{
		stderr:       stderr,
		outputBuffer: outputBuffer,
		tfDataDir:    tfDataDir,
	}
}

func (c Cmd) Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error {
	return c.RunWithEnv(stdout, workingDirectory, args, []string{}, debug)
}

func (c Cmd) RunWithEnv(stdout io.Writer, workingDirectory string, args []string, extraEnvVars []string, debug bool) error {
	command := exec.Command("terraform", args...)
	command.Dir = workingDirectory

	command.Env = os.Environ()
	command.Env = append(command.Env, extraEnvVars...)

	if debug {
		command.Stdout = io.MultiWriter(stdout, c.outputBuffer)
		command.Stderr = io.MultiWriter(c.stderr, c.outputBuffer)
	} else {
		command.Stdout = c.outputBuffer
		command.Stderr = c.outputBuffer
	}

	return command.Run()
}
