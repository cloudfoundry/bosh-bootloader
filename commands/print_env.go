package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type PrintEnv struct {
	stateValidator   stateValidator
	logger           logger
	stderrLogger     logger
	allProxyGetter   allProxyGetter
	terraformManager terraformManager
	credhubGetter    credhubGetter
	fs               fs
}

type envSetter interface {
	Set(key, value string) error
}

type credhubGetter interface {
	GetServer() (string, error)
	GetCerts() (string, error)
	GetPassword() (string, error)
}

type allProxyGetter interface {
	GeneratePrivateKey() (string, error)
	BoshAllProxy(string, string) string
}

type fs interface {
	fileio.TempDirer
	fileio.FileWriter
}

func NewPrintEnv(
	logger logger,
	stderrLogger logger,
	stateValidator stateValidator,
	allProxyGetter allProxyGetter,
	credhubGetter credhubGetter,
	terraformManager terraformManager,
	fs fs) PrintEnv {
	return PrintEnv{
		stateValidator:   stateValidator,
		logger:           logger,
		stderrLogger:     stderrLogger,
		allProxyGetter:   allProxyGetter,
		terraformManager: terraformManager,
		credhubGetter:    credhubGetter,
		fs:               fs,
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
		terraformOutputs, err := p.terraformManager.GetOutputs()
		if err != nil {
			return err
		}

		p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=https://%s:25555", terraformOutputs.GetString("external_ip")))
		return nil
	}

	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT=%s", state.BOSH.DirectorUsername))
	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT_SECRET=%s", state.BOSH.DirectorPassword))
	p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=%s", state.BOSH.DirectorAddress))
	p.logger.Println(fmt.Sprintf("export BOSH_CA_CERT='%s'", state.BOSH.DirectorSSLCA))

	p.logger.Println("export CREDHUB_USER=credhub-cli")

	credhubPassword, err := p.credhubGetter.GetPassword()
	if err == nil {
		p.logger.Println(fmt.Sprintf("export CREDHUB_PASSWORD=%s", credhubPassword))
	} else {
		p.stderrLogger.Println("No credhub password found.")
	}

	credhubServer, err := p.credhubGetter.GetServer()
	if err == nil {
		p.logger.Println(fmt.Sprintf("export CREDHUB_SERVER=%s", credhubServer))
	} else {
		p.stderrLogger.Println("No credhub server found.")
	}

	credhubCerts, err := p.credhubGetter.GetCerts()
	if err == nil {
		p.logger.Println(fmt.Sprintf("export CREDHUB_CA_CERT='%s'", credhubCerts))
	} else {
		p.stderrLogger.Println("No credhub certs found.")
	}

	privateKeyPath, err := p.allProxyGetter.GeneratePrivateKey()
	if err != nil {
		return err
	}

	p.logger.Println(fmt.Sprintf("export JUMPBOX_PRIVATE_KEY=%s", privateKeyPath))
	p.logger.Println(fmt.Sprintf("export BOSH_ALL_PROXY=%s", p.allProxyGetter.BoshAllProxy(state.Jumpbox.URL, privateKeyPath)))

	return nil
}
