package actors

import (
	"os"
	"os/exec"
)

type BOSHCLI struct{}

func NewBOSHCLI() BOSHCLI {
	return BOSHCLI{}
}

func (BOSHCLI) DirectorExists(address, username, password, caCertPath string) (bool, error) {
	cmd := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"-e", address,
		"--client", username,
		"--client-secret", password,
		"env",
	)
	cmd.Env = os.Environ()
	_, err := cmd.Output()

	return err == nil, err
}

func (BOSHCLI) Env(address, caCertPath string) (string, error) {
	cmd := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"-e", address,
		"env",
	)
	cmd.Env = os.Environ()
	env, err := cmd.Output()

	return string(env), err
}

func (b BOSHCLI) CloudConfig(address, caCertPath, username, password string) (string, error) {
	cmd := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"cloud-config",
	)
	cmd.Env = os.Environ()
	cloudConfig, err := cmd.Output()

	return string(cloudConfig), err
}

func (b BOSHCLI) RuntimeConfig(address, caCertPath, username, password, configName string) (string, error) {
	cmd := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"config",
		"--type", "runtime",
		"--name", configName,
	)
	cmd.Env = os.Environ()
	runtimeConfig, err := cmd.Output()

	return string(runtimeConfig), err
}
