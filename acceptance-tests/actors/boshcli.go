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

func (BOSHCLI) DeleteEnv(stateFilePath, manifestPath string) error {
	_, err := exec.Command(
		"bosh",
		"delete-env",
		fmt.Sprintf("--state=%s", stateFilePath),
		manifestPath,
	).Output()

	return err
}

func (BOSHCLI) Deploy(address, caCertPath, username, password, deployment, manifest, varsStore string, opsFiles []string, vars map[string]string) error {
	args := []string{
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"-d", deployment,
		"deploy", manifest,
		"--vars-store", varsStore,
		"-n",
	}
	for _, opsFile := range opsFiles {
		args = append(args, "-o", opsFile)
	}
	for key, value := range vars {
		args = append(args, "-v", fmt.Sprintf("%s=%s", key, value))
	}

	return exec.Command("bosh", args...).Run()
}

func (BOSHCLI) UploadRelease(address, caCertPath, username, password, releasePath string) error {
	err := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"upload-release", releasePath,
	).Run()
	if err != nil {
		fmt.Printf("bosh upload-release output: %s\n", string(err.(*exec.ExitError).Stderr))
	}
	return err
}

func (BOSHCLI) UploadStemcell(address, caCertPath, username, password, stemcellPath string) error {
	return exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"upload-stemcell", stemcellPath,
	).Run()
}

func (BOSHCLI) VMs(address, caCertPath, username, password, deployment string) (string, error) {
	output, err := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"-d", deployment,
		"vms",
	).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (BOSHCLI) DeleteDeployment(address, caCertPath, username, password, deployment string) error {
	return exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"-d", deployment,
		"delete-deployment",
		"-n",
	).Run()
}
