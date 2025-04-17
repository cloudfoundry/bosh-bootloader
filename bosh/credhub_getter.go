package bosh

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fileio"

	"gopkg.in/yaml.v2"
)

type CredhubGetter struct {
	stateStore stateStore
	reader     fileio.FileReader
}

func NewCredhubGetter(stateStore stateStore, fileIO fileio.FileReader) CredhubGetter {
	return CredhubGetter{
		stateStore: stateStore,
		reader:     fileIO,
	}
}

func (c CredhubGetter) GetServer() (string, error) {
	var p struct {
		InternalIp string `yaml:"internal_ip"`
	}

	varsDir, err := c.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars directory: %s", err) //nolint:staticcheck
	}

	varsFile, err := c.reader.ReadFile(filepath.Join(varsDir, "director-vars-file.yml"))
	if err != nil {
		return "", fmt.Errorf("Read director-vars-file.yml file: %s", err) //nolint:staticcheck
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
		return "", fmt.Errorf("Get vars directory: %s", err) //nolint:staticcheck
	}

	varsStore, err := c.reader.ReadFile(filepath.Join(varsDir, "director-vars-store.yml"))
	if err != nil {
		return "", fmt.Errorf("Read director-vars-store.yml file: %s", err) //nolint:staticcheck
	}

	err = yaml.Unmarshal(varsStore, &certs)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", certs.CredhubCA.Certificate, certs.UAASSL.Certificate), nil
}

func (c CredhubGetter) GetPassword() (string, error) {
	var certs struct {
		Password string `yaml:"credhub_admin_client_secret"`
	}

	varsDir, err := c.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars directory: %s", err) //nolint:staticcheck
	}

	varsStore, err := c.reader.ReadFile(filepath.Join(varsDir, "director-vars-store.yml"))
	if err != nil {
		return "", fmt.Errorf("Read director-vars-store.yml file: %s", err) //nolint:staticcheck
	}

	err = yaml.Unmarshal(varsStore, &certs)
	if err != nil {
		return "", err
	}

	return certs.Password, nil
}
