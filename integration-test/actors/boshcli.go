package actors

import (
	"fmt"
	"os/exec"
)

type BOSHCLI struct{}

func NewBOSHCLI() BOSHCLI {
	return BOSHCLI{}
}

func (BOSHCLI) DirectorExists(address, caCertPath string) (bool, error) {
	_, err := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"-e", address,
		"env",
	).Output()

	return err == nil, err
}

func (BOSHCLI) DeleteEnv(stateFilePath, manifestPath string) error {
	_, err := exec.Command(
		"bosh",
		"delete-env",
		fmt.Sprintf("--state=%s", stateFilePath),
		manifestPath,
	).Output()

	return err
}
