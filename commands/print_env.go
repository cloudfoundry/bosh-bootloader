package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

// PrintEnv defines a PrintEnv command
type PrintEnv struct {
	stateValidator   stateValidator
	logger           logger
	stderrLogger     logger
	allProxyGetter   allProxyGetter
	terraformManager terraformManager
	credhubGetter    credhubGetter
	fs               fs
	rendererFactory  renderers.Factory
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

type PrintEnvConfig struct {
	shellType string
}

// NewPrintEnv creates a new PrintEnv Command
func NewPrintEnv(
	logger logger,
	stderrLogger logger,
	stateValidator stateValidator,
	allProxyGetter allProxyGetter,
	credhubGetter credhubGetter,
	terraformManager terraformManager,
	fs fs,
	rendererFactory renderers.Factory) PrintEnv {
	return PrintEnv{
		stateValidator:   stateValidator,
		logger:           logger,
		stderrLogger:     stderrLogger,
		allProxyGetter:   allProxyGetter,
		terraformManager: terraformManager,
		credhubGetter:    credhubGetter,
		fs:               fs,
		rendererFactory:  rendererFactory,
	}
}

func (p PrintEnv) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := p.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (p PrintEnv) ParseArgs(args []string, state storage.State) (PrintEnvConfig, error) {
	var (
		config PrintEnvConfig
	)

	printEnvFlags := flags.New("print-env")
	printEnvFlags.String(&config.shellType, "shell-type", "")

	err := printEnvFlags.Parse(args)
	if err != nil {
		return PrintEnvConfig{}, err
	}

	return config, nil
}

func (p PrintEnv) Execute(args []string, state storage.State) error {
	variables := make(map[string]string)

	config, err := p.ParseArgs(args, state)
	if err != nil {
		return err
	}

	shell := config.shellType
	renderer, err := p.rendererFactory.Create(shell)
	if err != nil {
		return err
	}

	variables["BOSH_CLIENT"] = state.BOSH.DirectorUsername
	variables["BOSH_CLIENT_SECRET"] = state.BOSH.DirectorPassword
	variables["BOSH_ENVIRONMENT"] = state.BOSH.DirectorAddress
	variables["BOSH_CA_CERT"] = state.BOSH.DirectorSSLCA
	variables["CREDHUB_CLIENT"] = "credhub-admin"

	credhubPassword, err := p.credhubGetter.GetPassword()
	if err == nil {
		variables["CREDHUB_SECRET"] = credhubPassword
	} else {
		p.stderrLogger.Println("No credhub password found.")
	}

	credhubServer, err := p.credhubGetter.GetServer()
	if err == nil {
		variables["CREDHUB_SERVER"] = credhubServer
	} else {
		p.stderrLogger.Println("No credhub server found.")
	}

	credhubCerts, err := p.credhubGetter.GetCerts()
	if err == nil {
		variables["CREDHUB_CA_CERT"] = credhubCerts
	} else {
		p.stderrLogger.Println("No credhub certs found.")
	}

	privateKeyPath, err := p.allProxyGetter.GeneratePrivateKey()
	if err != nil {
		p.renderVariables(renderer, variables)
		return err
	}

	variables["JUMPBOX_PRIVATE_KEY"] = privateKeyPath
	variables["BOSH_ALL_PROXY"] = p.allProxyGetter.BoshAllProxy(state.Jumpbox.URL, privateKeyPath)
	variables["CREDHUB_PROXY"] = p.allProxyGetter.BoshAllProxy(state.Jumpbox.URL, privateKeyPath)

	p.renderVariables(renderer, variables)
	return nil
}

func (p PrintEnv) renderVariables(renderer renderers.Renderer, variables map[string]string) {
	for k, v := range variables {
		p.logger.Println(renderer.RenderEnvironmentVariable(k, v))
	}
}
