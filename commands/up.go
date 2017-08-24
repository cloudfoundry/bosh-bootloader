package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Up struct {
	awsUp       awsUp
	azureUp     azureUp
	gcpUp       gcpUp
	envGetter   envGetter
	boshManager boshManager
}

type awsUp interface {
	Execute(awsUpConfig AWSUpConfig, state storage.State) error
}

type gcpUp interface {
	Execute(gcpUpConfig GCPUpConfig, state storage.State) error
}

type azureUp interface {
	Execute(azureUpConfig AzureUpConfig, state storage.State) error
}

type envGetter interface {
	Get(name string) string
}

type upConfig struct {
	name       string
	opsFile    string
	noDirector bool
	jumpbox    bool
}

func NewUp(awsUp awsUp, gcpUp gcpUp, azureUp azureUp, envGetter envGetter, boshManager boshManager) Up {
	return Up{
		awsUp:       awsUp,
		azureUp:     azureUp,
		gcpUp:       gcpUp,
		envGetter:   envGetter,
		boshManager: boshManager,
	}
}

func (u Up) CheckFastFails(args []string, state storage.State) error {
	config, err := u.parseArgs(state, args)
	if err != nil {
		return err
	}

	if !config.noDirector && !state.NoDirector {
		err = fastFailBOSHVersion(u.boshManager)
		if err != nil {
			return err
		}
	}

	if config.jumpbox && !state.Jumpbox.Enabled && state.EnvID != "" {
		return errors.New(`Environment without credhub already exists, you must recreate your environment to use "--credhub"`)
	}

	if state.EnvID != "" && config.name != "" && config.name != state.EnvID {
		return fmt.Errorf("The director name cannot be changed for an existing environment. Current name is %s.", state.EnvID)
	}

	return nil
}

func (u Up) Execute(args []string, state storage.State) error {
	config, err := u.parseArgs(state, args)
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "aws":
		err = u.awsUp.Execute(AWSUpConfig{
			OpsFilePath: config.opsFile,
			Name:        config.name,
			NoDirector:  config.noDirector,
			Jumpbox:     config.jumpbox,
		}, state)
	case "gcp":
		err = u.gcpUp.Execute(GCPUpConfig{
			OpsFilePath: config.opsFile,
			Name:        config.name,
			NoDirector:  config.noDirector,
			Jumpbox:     config.jumpbox,
		}, state)
	case "azure":
		err = u.azureUp.Execute(AzureUpConfig{
			Name:       config.name,
			NoDirector: config.noDirector,
		}, state)
	}

	if err != nil {
		return err
	}

	return nil
}

func (u Up) parseArgs(state storage.State, args []string) (upConfig, error) {
	var config upConfig

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return upConfig{}, err //not tested
	}

	prevOpsFilePath := filepath.Join(tempDir, "user-ops-file")
	err = ioutil.WriteFile(prevOpsFilePath, []byte(state.BOSH.UserOpsFile), os.ModePerm)
	if err != nil {
		return upConfig{}, err //not tested
	}

	upFlags := flags.New("up")

	upFlags.String(&config.name, "name", "")
	upFlags.String(&config.opsFile, "ops-file", prevOpsFilePath)
	upFlags.Bool(&config.noDirector, "", "no-director", state.NoDirector)
	upFlags.Bool(&config.jumpbox, "", "credhub", state.Jumpbox.Enabled)

	err = upFlags.Parse(args)
	if err != nil {
		return upConfig{}, err
	}

	return config, nil
}
