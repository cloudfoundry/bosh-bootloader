package actors

import "os/exec"

type BOSHCLI struct{}

func NewBOSHCLI() BOSHCLI {
	return BOSHCLI{}
}

func (BOSHCLI) DirectorExists(address, caCertPath string) (bool, error) {
	_, err := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"env", address,
	).Output()
	return err == nil, err
}
