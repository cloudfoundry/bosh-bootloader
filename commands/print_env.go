package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	PrintEnvCommand = "print-env"
)

type PrintEnv struct {
	stateValidator stateValidator
	logger         logger
}

type envSetter interface {
	Set(key, value string) error
}

func NewPrintEnv(logger logger, stateValidator stateValidator) PrintEnv {
	return PrintEnv{
		stateValidator: stateValidator,
		logger:         logger,
	}
}

func (p PrintEnv) Execute(args []string, state storage.State) error {
	err := p.stateValidator.Validate()
	if err != nil {
		return err
	}

	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT=%s", state.BOSH.DirectorUsername))
	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT_SECRET=%s", state.BOSH.DirectorPassword))
	p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=%s", state.BOSH.DirectorAddress))
	p.logger.Println(fmt.Sprintf("export BOSH_CA_CERT='%s'", state.BOSH.DirectorSSLCA))

	return nil
}
