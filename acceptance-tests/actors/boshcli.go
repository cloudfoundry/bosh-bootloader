package actors

import (
	"os/exec"
)

type BOSHCLI struct{}

func NewBOSHCLI() BOSHCLI {
	return BOSHCLI{}
}

func (BOSHCLI) DirectorExists(address, username, password, caCertPath string) (bool, error) {
	_, err := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"-e", address,
		"--client", username,
		"--client-secret", password,
		"env",
	).Output()

	return err == nil, err
}

func (BOSHCLI) Env(address, caCertPath string) (string, error) {
	env, err := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"-e", address,
		"env",
	).Output()

	return string(env), err
}

func (BOSHCLI) CloudConfig(address, caCertPath, username, password string) (string, error) {
	cloudConfig, err := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"cloud-config",
	).Output()

	return string(cloudConfig), err
}
