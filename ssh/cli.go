package ssh

import (
	"io"
	"os/exec"
)

type CLI struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func NewCLI(in io.Reader, out, err io.Writer) CLI {
	return CLI{
		in:  in,
		out: out,
		err: err,
	}
}

func (c CLI) Run(args []string) error {
	command := exec.Command("ssh", args...)

	command.Stdin = c.in
	command.Stdout = c.out
	command.Stderr = c.err

	return command.Run()
}
