package bosh

import (
	"io"
	"os"
	"os/exec"
)

type AuthenticatedCLI struct {
	GlobalArgs         []string
	BOSHAllProxy       string
	Stderr             io.Writer
	BOSHExecutablePath string
}

func NewAuthenticatedCLI(stderr io.Writer, boshPath, directorAddress, username, password, caCert, boshAllProxy string) AuthenticatedCLI {
	return AuthenticatedCLI{
		GlobalArgs: []string{
			"--environment", directorAddress,
			"--client", username,
			"--client-secret", password,
			"--ca-cert", caCert,
			"--non-interactive",
		},
		BOSHAllProxy:       boshAllProxy,
		Stderr:             stderr,
		BOSHExecutablePath: boshPath,
	}
}

func (c AuthenticatedCLI) Run(stdout io.Writer, workingDirectory string, args []string) error {
	command := exec.Command(c.BOSHExecutablePath, append(c.GlobalArgs, args...)...)
	command.Env = append(os.Environ(), "BOSH_ALL_PROXY="+c.BOSHAllProxy)
	command.Stdout = stdout
	command.Stderr = c.Stderr

	return command.Run()
}
