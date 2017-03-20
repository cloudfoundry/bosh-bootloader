package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	PrintEnvCommand = "print-env"
)

type PrintEnv struct {
	stateValidator        stateValidator
	logger                logger
	terraformManager      terraformManager
	infrastructureManager infrastructureManager
}

type envSetter interface {
	Set(key, value string) error
}

func NewPrintEnv(logger logger, stateValidator stateValidator, terraformManager terraformManager, infrastructureManager infrastructureManager) PrintEnv {
	return PrintEnv{
		stateValidator:        stateValidator,
		logger:                logger,
		terraformManager:      terraformManager,
		infrastructureManager: infrastructureManager,
	}
}

func (p PrintEnv) Execute(args []string, state storage.State) error {
	err := p.stateValidator.Validate()
	if err != nil {
		return err
	}

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
	switch state.IAAS {
	case "aws":
		stack, err := p.infrastructureManager.Describe(state.Stack.Name)
		if err != nil {
			return "", err
		}
		return stack.Outputs["BOSHEIP"], nil
	case "gcp":
		terraformOutputs, err := p.terraformManager.GetOutputs(state.TFState, state.LB.Type, false)
		if err != nil {
			return "", err
		}
		return terraformOutputs.ExternalIP, nil
	}

	return "", errors.New("Could not find external IP for given IAAS")
}
