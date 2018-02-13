package terraform

import (
	"io"
	"os/exec"
)

type Cmd struct {
	stderr       io.Writer
	outputBuffer io.Writer
}

func NewCmd(stderr, outputBuffer io.Writer) Cmd {
	return Cmd{
		stderr:       stderr,
		outputBuffer: outputBuffer,
	}
}

func (c Cmd) Run(stdout io.Writer, args []string, debug bool) error {
	command := exec.Command("terraform", args...)

	if debug {
		command.Stdout = io.MultiWriter(stdout, c.outputBuffer)
		command.Stderr = io.MultiWriter(c.stderr, c.outputBuffer)
	} else {
		command.Stdout = c.outputBuffer
		command.Stderr = c.outputBuffer
	}

	return command.Run()
}
