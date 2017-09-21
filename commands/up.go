package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Up struct {
	upCmd       UpCmd
	boshManager boshManager
}

type UpCmd interface {
	Execute(upConfig UpConfig, state storage.State) error
}

type UpConfig struct {
	Name       string
	OpsFile    string
	NoDirector bool
}

func NewUp(upCmd UpCmd, boshManager boshManager) Up {
	return Up{
		upCmd:       upCmd,
		boshManager: boshManager,
	}
}

func (u Up) CheckFastFails(args []string, state storage.State) error {
	config, err := u.parseArgs(state, args)
	if err != nil {
		return err
	}

	if !config.NoDirector && !state.NoDirector {
		err = fastFailBOSHVersion(u.boshManager)
		if err != nil {
			return err
		}
	}

	if state.EnvID != "" && config.Name != "" && config.Name != state.EnvID {
		return fmt.Errorf("The director name cannot be changed for an existing environment. Current name is %s.", state.EnvID)
	}

	return nil
}

func (u Up) Execute(args []string, state storage.State) error {
	config, err := u.parseArgs(state, args)
	if err != nil {
		return err
	}

	return u.upCmd.Execute(UpConfig{
		OpsFile:    config.OpsFile,
		Name:       config.Name,
		NoDirector: config.NoDirector,
	}, state)
}

func (u Up) parseArgs(state storage.State, args []string) (UpConfig, error) {
	var config UpConfig

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return UpConfig{}, err //not tested
	}

	prevOpsFilePath := filepath.Join(tempDir, "user-ops-file")
	err = ioutil.WriteFile(prevOpsFilePath, []byte(state.BOSH.UserOpsFile), os.ModePerm)
	if err != nil {
		return UpConfig{}, err //not tested
	}

	upFlags := flags.New("up")

	upFlags.String(&config.Name, "name", "")
	upFlags.String(&config.OpsFile, "ops-file", prevOpsFilePath)
	upFlags.Bool(&config.NoDirector, "", "no-director", state.NoDirector)

	err = upFlags.Parse(args)
	if err != nil {
		return UpConfig{}, err
	}

	return config, nil
}
