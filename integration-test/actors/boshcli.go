package actors

import "os/exec"

type BOSHCLI struct{}

func NewBOSHCLI() BOSHCLI {
	return BOSHCLI{}
}

func (BOSHCLI) DirectorExists(address, caCertPath string) bool {
	_, err := exec.Command("bosh",
		"env", address,
		"-c", caCertPath).Output()
	return err == nil
}
