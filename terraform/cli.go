package terraform

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type CLI struct {
	errorBuffer     io.Writer
	outputBuffer    io.Writer
	tfDataDir       string
	terraformBinary string
}

func NewCLI(errorBuffer, outputBuffer io.Writer, tfDataDir string, terraformBinary string) CLI {
	return CLI{
		errorBuffer:     errorBuffer,
		outputBuffer:    outputBuffer,
		tfDataDir:       tfDataDir,
		terraformBinary: terraformBinary,
	}
}

func (c CLI) Run(stdout io.Writer, workingDirectory string, args []string) error {
	return c.RunWithEnv(stdout, workingDirectory, args, []string{})
}

func (c CLI) RunWithEnv(stdout io.Writer, workingDirectory string, args []string, extraEnvVars []string) error {
	path, err := NewBinary(c.terraformBinary).BinaryPath()
	if err != nil {
		return err
	}
	command := exec.Command(path, args...)
	command.Dir = workingDirectory

	command.Env = os.Environ()
	command.Env = append(command.Env, extraEnvVars...)

	command.Stdout = io.MultiWriter(stdout, c.outputBuffer)
	command.Stderr = c.errorBuffer

	err = command.Run()
	if err != nil {
		return fmt.Errorf("command execution failed got: %s stderr:\n %s", err, c.errorBuffer)
	}

	return nil
}
