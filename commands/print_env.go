package commands

import (
	"fmt"

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
	if !state.NoDirector {
		p.logger.Println(fmt.Sprintf("export BOSH_CLIENT=%s", state.BOSH.DirectorUsername))
		p.logger.Println(fmt.Sprintf("export BOSH_CLIENT_SECRET=%s", state.BOSH.DirectorPassword))
		p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=%s", state.BOSH.DirectorAddress))
		p.logger.Println(fmt.Sprintf("export BOSH_CA_CERT='%s'", state.BOSH.DirectorSSLCA))
	} else {
		directorAddress, err := p.getExternalIP(state)
		if err != nil {
			return err
		}
		p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=https://%s:25555", directorAddress))
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
