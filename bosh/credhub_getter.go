package bosh

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type CredhubGetter struct {
	stateStore stateStore
}

func NewCredhubGetter(stateStore stateStore) CredhubGetter {
	return CredhubGetter{
		stateStore: stateStore,
	}
}

func (c CredhubGetter) GetServer() (string, error) {
	var p struct {
		InternalIp string `yaml:"internal_ip"`
	}

	varsDir, err := c.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars directory: %s", err)
	}

	varsFile, err := ioutil.ReadFile(filepath.Join(varsDir, "director-vars-file.yml"))
	if err != nil {
		return "", fmt.Errorf("Read director-vars-file.yml file: %s", err)
	}

	err = yaml.Unmarshal(varsFile, &p)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s:8844", p.InternalIp), nil
}

func (c CredhubGetter) GetCerts() (string, error) {
	var certs struct {
		CredhubCA struct {
			Certificate string `yaml:"certificate"`
		} `yaml:"credhub_ca"`
		UAASSL struct {
			Certificate string `yaml:"certificate"`
		} `yaml:"uaa_ssl"`
	}

	varsDir, err := c.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars directory: %s", err)
	}

	varsStore, err := ioutil.ReadFile(filepath.Join(varsDir, "director-vars-store.yml"))
	if err != nil {
		return "", fmt.Errorf("Read director-vars-store.yml file: %s", err)
	}

	err = yaml.Unmarshal(varsStore, &certs)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", certs.CredhubCA.Certificate, certs.UAASSL.Certificate), nil
}

func (c CredhubGetter) GetPassword() (string, error) {
	var certs struct {
		Password string `yaml:"credhub_cli_password"`
	}

	varsDir, err := c.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars directory: %s", err)
	}

	varsStore, err := ioutil.ReadFile(filepath.Join(varsDir, "director-vars-store.yml"))
	if err != nil {
		return "", fmt.Errorf("Read director-vars-store.yml file: %s", err)
	}

	err = yaml.Unmarshal(varsStore, &certs)
	if err != nil {
		return "", err
	}

	return certs.Password, nil
}
