package actors

import (
	"encoding/json"
	"errors"
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

func (b BOSHCLI) UploadStemcell(address, caCertPath, username, password, stemcellURL string) error {
	cmd := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"upload-stemcell", stemcellURL,
	)

	cmd.Env = os.Environ()
	_, err := cmd.Output()

	return err
}

// Stemcells returns a list of cid's for the uploaded stemcells
func (b BOSHCLI) Stemcells(address, caCertPath, username, password string) ([]string, error) {
	var cids []string

	cmd := exec.Command("bosh",
		"--ca-cert", caCertPath,
		"--client", username,
		"--client-secret", password,
		"-e", address,
		"stemcells", "--json",
	)

	cmd.Env = os.Environ()
	data, err := cmd.Output()
	if err != nil {
		return cids, errors.New("could not unmarshal response: " + err.Error())
	}

	body := struct {
		Tables []struct {
			Rows []struct {
				CID string `json:"cid"`
			}
		}
	}{}

	if err := json.Unmarshal(data, &body); err != nil {
		return cids, errors.New("could not unmarshal response: " + err.Error())
	}

	if len(body.Tables) > 0 {
		for _, row := range body.Tables[0].Rows {
			cids = append(cids, row.CID)
		}
	}

	return cids, nil
}
