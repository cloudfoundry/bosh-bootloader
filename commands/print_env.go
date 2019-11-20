package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

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
	shellType    string
	metadataFile string
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
	// We don't do any validation here, because at this point, we don't know if we're using
	// a bbl-state or a metadata file. Once we know that we're using a bbl-state, in Execute,
	// we validate that the state exists.
	return nil
}

func (p PrintEnv) ParseArgs(args []string, state storage.State) (PrintEnvConfig, error) {
	var (
		config PrintEnvConfig
	)

	printEnvFlags := flags.New("print-env")
	printEnvFlags.String(&config.shellType, "shell-type", "")
	printEnvFlags.String(&config.metadataFile, "metadata-file", "")

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

	metadataFile := config.metadataFile
	if metadataFile != "" {
		var metadata struct {
			Name     string            `json:"name"`
			IaasType string            `json:"iaas_type"`
			Bosh     map[string]string `json:"bosh"`
		}
		metadataContents, err := ioutil.ReadFile(metadataFile)
		if err != nil {
			p.stderrLogger.Println(fmt.Sprintf("Failed to read %s: %s", metadataFile, err))
			return err
		}

		err = json.Unmarshal(metadataContents, &metadata)
		if err != nil {
			p.stderrLogger.Println(fmt.Sprintf("Failed to unmarshal %s: %s", metadataFile, err))
			return err
		}

		variables["BOSH_CLIENT"] = metadata.Bosh["bosh_client"]
		variables["BOSH_CLIENT_SECRET"] = metadata.Bosh["bosh_client_secret"]
		variables["BOSH_ENVIRONMENT"] = metadata.Bosh["bosh_environment"]
		variables["BOSH_CA_CERT"] = metadata.Bosh["bosh_ca_cert"]
		variables["CREDHUB_CLIENT"] = metadata.Bosh["credhub_client"]
		variables["CREDHUB_SECRET"] = metadata.Bosh["credhub_secret"]
		variables["CREDHUB_SERVER"] = metadata.Bosh["credhub_server"]
		variables["CREDHUB_CA_CERT"] = metadata.Bosh["credhub_ca_cert"]

		privateKeyPath := fmt.Sprintf("/tmp/%s.priv", metadata.Name)
		err = ioutil.WriteFile(privateKeyPath, []byte(metadata.Bosh["jumpbox_private_key"]), 0600)
		if err != nil {
			p.stderrLogger.Println(fmt.Sprintf("Failed to write private key to %s: %s", privateKeyPath, err))
			return err
		}

		variables["JUMPBOX_PRIVATE_KEY"] = privateKeyPath
		boshAllProxy := fmt.Sprintf("%s=%s", strings.Split(metadata.Bosh["bosh_all_proxy"], "=")[0], privateKeyPath)
		variables["BOSH_ALL_PROXY"] = boshAllProxy
		variables["CREDHUB_PROXY"] = boshAllProxy

		p.renderVariables(renderer, variables)
		return nil
	}

	err = p.stateValidator.Validate()
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
	variables["BOSH_ALL_PROXY"] = p.allProxyGetter.BoshAllProxy(state.Jumpbox.GetURLWithJumpboxUser(), privateKeyPath)
	variables["CREDHUB_PROXY"] = p.allProxyGetter.BoshAllProxy(state.Jumpbox.GetURLWithJumpboxUser(), privateKeyPath)

	p.renderVariables(renderer, variables)
	return nil
}

func (p PrintEnv) renderVariables(renderer renderers.Renderer, variables map[string]string) {
	for k, v := range variables {
		p.logger.Println(renderer.RenderEnvironmentVariable(k, v))
	}
}
