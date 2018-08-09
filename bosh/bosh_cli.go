package bosh

import (
	"io"
	"os"
	"os/exec"
)

type BOSHCLI struct {
	GlobalArgs   []string
	BOSHAllProxy string
	Stderr       io.Writer
	BOSHCLIPath  string
}

func NewBOSHCLI(stderr io.Writer, boshPath, directorAddress, username, password, caCert, boshAllProxy string) BOSHCLI {
	return BOSHCLI{
		GlobalArgs: []string{
			"--environment", directorAddress,
			"--client", username,
			"--client-secret", password,
			"--ca-cert", caCert,
			"--non-interactive",
		},
		BOSHAllProxy: boshAllProxy,
		Stderr:       stderr,
		BOSHCLIPath:  boshPath,
	}
}

func (c BOSHCLI) UpdateRuntimeConfig(filepath, name string) error {
	args := []string{
		"update-runtime-config", filepath,
		"--name", name,
	}
	return c.Run(nil, "", args)
}

func (c BOSHCLI) Run(stdout io.Writer, workingDirectory string, args []string) error {
	command := exec.Command(c.BOSHCLIPath, append(c.GlobalArgs, args...)...)
	command.Env = append(os.Environ(), "BOSH_ALL_PROXY="+c.BOSHAllProxy)
	command.Stdout = stdout
	command.Stderr = c.Stderr

	return command.Run()
}
