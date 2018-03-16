package ssh

import (
	"fmt"
	"io"
	"os/exec"
)

type Cmd struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func NewCmd(in io.Reader, out, err io.Writer) Cmd {
	return Cmd{
		in:  in,
		out: out,
		err: err,
	}
}

func (c Cmd) Run(args []string) error {
	command := exec.Command("ssh", args...)

	fmt.Printf("ssh %+v\n", args)

	command.Stdin = c.in
	command.Stdout = c.out
	command.Stderr = c.err

	return command.Run()
}
