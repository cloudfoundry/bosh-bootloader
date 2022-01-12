package terraform

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type CLI struct {
	errorBuffer  io.Writer
	outputBuffer io.Writer
	tfDataDir    string
}

func NewCLI(errorBuffer, outputBuffer io.Writer, tfDataDir string) CLI {
	return CLI{
		errorBuffer:  errorBuffer,
		outputBuffer: outputBuffer,
		tfDataDir:    tfDataDir,
	}
}

func (c CLI) Run(stdout io.Writer, workingDirectory string, args []string) error {
	return c.RunWithEnv(stdout, workingDirectory, args, []string{})
}

func (c CLI) RunWithEnv(stdout io.Writer, workingDirectory string, args []string, extraEnvVars []string) error {
	path, err := NewBinary().BinaryPath()
	if err != nil {
		return err
	}
	if c.errorBuffer != nil {
		c.errorBuffer.Write([]byte(fmt.Sprintf("Use terraform from path '%s'.\n", path)))
	}

	command := exec.Command(path, args...)
	command.Dir = workingDirectory

	command.Env = os.Environ()
	command.Env = append(command.Env, extraEnvVars...)

	command.Stdout = io.MultiWriter(stdout, c.outputBuffer)
	command.Stderr = c.errorBuffer
	command.Stdin = os.Stdin

	return command.Run()
}
