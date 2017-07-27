package commands

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	PrintEnvCommand = "print-env"
)

type PrintEnv struct {
	stateValidator   stateValidator
	logger           logger
	terraformManager terraformOutputter
}

type envSetter interface {
	Set(key, value string) error
}

func NewPrintEnv(logger logger, stateValidator stateValidator, terraformManager terraformOutputter) PrintEnv {
	return PrintEnv{
		stateValidator:   stateValidator,
		logger:           logger,
		terraformManager: terraformManager,
	}
}

func (p PrintEnv) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := p.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (p PrintEnv) Execute(args []string, state storage.State) error {
	if state.NoDirector {
		directorAddress, err := p.getExternalIP(state)
		if err != nil {
			return err
		}
		p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=https://%s:25555", directorAddress))

		return nil
	}

	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT=%s", state.BOSH.DirectorUsername))
	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT_SECRET=%s", state.BOSH.DirectorPassword))
	p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=%s", state.BOSH.DirectorAddress))
	p.logger.Println(fmt.Sprintf("export BOSH_CA_CERT='%s'", state.BOSH.DirectorSSLCA))

	if state.Jumpbox.Enabled {
		portNumber, err := p.getPort()
		if err != nil {
			// not tested
			return err
		}

		dir, err := ioutil.TempDir("", "bosh-jumpbox")
		if err != nil {
			// not tested
			return err
		}

		privateKeyPath := filepath.Join(dir, "bosh_jumpbox_private.key")

		privateKeyContents, err := p.privateKeyFromJumpboxVariables(state.Jumpbox.Variables)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(privateKeyPath, []byte(privateKeyContents), 0600)
		if err != nil {
			// not tested
			return err
		}

		jumpboxURL := strings.Split(state.Jumpbox.URL, ":")[0]

		p.logger.Println(fmt.Sprintf("export BOSH_ALL_PROXY=socks5://localhost:%s", portNumber))
		p.logger.Println(fmt.Sprintf("export BOSH_GW_PRIVATE_KEY=%s", privateKeyPath))
		p.logger.Println(fmt.Sprintf("ssh -f -N -D %s jumpbox@%s -i $BOSH_GW_PRIVATE_KEY", portNumber, jumpboxURL))
	}

	return nil
}

func (p PrintEnv) getExternalIP(state storage.State) (string, error) {
	terraformOutputs, err := p.terraformManager.GetOutputs(state)
	if err != nil {
		return "", err
	}

	return terraformOutputs["external_ip"].(string), nil
}

func (p PrintEnv) getPort() (string, error) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	defer l.Close()

	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return "", err
	}

	return port, nil
}

func (p PrintEnv) privateKeyFromJumpboxVariables(jumpboxVariables string) (string, error) {
	var jumpboxVars struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	err := yaml.Unmarshal([]byte(jumpboxVariables), &jumpboxVars)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling jumpbox variables: %v", err)
	}

	return jumpboxVars.JumpboxSSH.PrivateKey, nil
}
