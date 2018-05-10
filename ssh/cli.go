package ssh

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
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

// background execute
func (c CLI) Start(args []string) (*exec.Cmd, error) {
	fmt.Fprintf(c.out, "starting:\nssh %s\n", strings.Join(args, " "))
	return c.start(args)
}

// foreground execute
func (c CLI) Run(args []string) error {
	fmt.Fprintf(c.out, "running:\nssh %s\n", strings.Join(args, " "))
	cmd, err := c.start(args)
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func (c CLI) start(args []string) (*exec.Cmd, error) {
	command := exec.Command("ssh", args...)

	command.Stdin = c.in
	command.Stdout = c.out
	command.Stderr = c.err

	return command, command.Start()
}
