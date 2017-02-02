package boshinit

import (
	"io"
	"os/exec"
	"path/filepath"
)

type CommandBuilder struct {
	Path      string
	Directory string
	Stdout    io.Writer
	Stderr    io.Writer
}

func NewCommandBuilder(path string, dir string, stdout io.Writer, stderr io.Writer) CommandBuilder {
	return CommandBuilder{
		Path:      path,
		Directory: dir,
		Stdout:    stdout,
		Stderr:    stderr,
	}
}

func (b CommandBuilder) DeployCommand() *exec.Cmd {
	return &exec.Cmd{
		Path: b.Path,
		Args: []string{
			filepath.Base(b.Path),
			"deploy",
			"bosh.yml",
		},
		Dir:    b.Directory,
		Stdout: b.Stdout,
		Stderr: b.Stderr,
	}
}

func (b CommandBuilder) DeleteCommand() *exec.Cmd {
	return &exec.Cmd{
		Path: b.Path,
		Args: []string{
			filepath.Base(b.Path),
			"delete",
			"bosh.yml",
		},
		Dir:    b.Directory,
		Stdout: b.Stdout,
		Stderr: b.Stderr,
	}
}
