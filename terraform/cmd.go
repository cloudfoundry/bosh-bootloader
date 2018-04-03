package terraform

import (
	"io"
	"os"
	"os/exec"
)

type Cmd struct {
	errorBuffer  io.Writer
	outputBuffer io.Writer
	tfDataDir    string
}

func NewCmd(errorBuffer, outputBuffer io.Writer, tfDataDir string) Cmd {
	return Cmd{
		errorBuffer:  errorBuffer,
		outputBuffer: outputBuffer,
		tfDataDir:    tfDataDir,
	}
}

func (c Cmd) Run(stdout io.Writer, workingDirectory string, args []string) error {
	return c.RunWithEnv(stdout, workingDirectory, args, []string{})
}

func (c Cmd) RunWithEnv(stdout io.Writer, workingDirectory string, args []string, extraEnvVars []string) error {
	command := exec.Command("terraform", args...)
	command.Dir = workingDirectory

	command.Env = os.Environ()
	command.Env = append(command.Env, extraEnvVars...)

	command.Stdout = io.MultiWriter(stdout, c.outputBuffer)
	command.Stderr = c.errorBuffer

	return command.Run()
}
