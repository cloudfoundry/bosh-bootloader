package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Up struct {
	upCmd              UpCmd
	boshManager        boshManager
	cloudConfigManager cloudConfigManager
	stateStore         stateStore
	envIDManager       envIDManager
	terraformManager   terraformApplier
}

type UpCmd interface {
	Execute(storage.State) (storage.State, error)
}

type UpConfig struct {
	Name       string
	OpsFile    string
	NoDirector bool
}

func NewUp(upCmd UpCmd, boshManager boshManager, cloudConfigManager cloudConfigManager,
	stateStore stateStore, envIDManager envIDManager, terraformManager terraformApplier) Up {
	return Up{
		upCmd:              upCmd,
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		stateStore:         stateStore,
		envIDManager:       envIDManager,
		terraformManager:   terraformManager,
	}
}

func (u Up) CheckFastFails(args []string, state storage.State) error {
	config, err := u.ParseArgs(args, state)
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
	err := u.terraformManager.ValidateVersion()
	if err != nil {
		return fmt.Errorf("Terraform validate version: %s", err)
	}

	state, err = u.upCmd.Execute(state)
	if err != nil {
		return err
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after IAAS up: %s", err)
	}

	config, err := u.ParseArgs(args, state)
	if err != nil {
		return err
	}

	if config.NoDirector {
		if !state.BOSH.IsEmpty() {
			return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}

		state.NoDirector = true
	}

	var opsFileContents []byte
	if config.OpsFile != "" {
		opsFileContents, err = ioutil.ReadFile(config.OpsFile)
		if err != nil {
			return fmt.Errorf("Reading ops-file contents: %v", err)
		}
	}

	state, err = u.envIDManager.Sync(state, config.Name)
	if err != nil {
		return fmt.Errorf("Env id manager sync: %s", err)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after sync: %s", err)
	}

	state, err = u.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, u.stateStore)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after terraform apply: %s", err)
	}

	if state.NoDirector {
		return nil
	}

	terraformOutputs, err := u.terraformManager.GetOutputs(state)
	if err != nil {
		return fmt.Errorf("Parse terraform outputs: %s", err)
	}

	state, err = u.boshManager.CreateJumpbox(state, terraformOutputs)
	if err != nil {
		return fmt.Errorf("Create jumpbox: %s", err)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after create jumpbox: %s", err)
	}

	state.BOSH.UserOpsFile = string(opsFileContents)
	state, err = u.boshManager.CreateDirector(state, terraformOutputs)
	switch err.(type) {
	case bosh.ManagerCreateError:
		bcErr := err.(bosh.ManagerCreateError)
		if setErr := u.stateStore.Set(bcErr.State()); setErr != nil {
			return fmt.Errorf("Save state after bosh director create error: %s, %s", err, setErr)
		}
		return fmt.Errorf("Create bosh director: %s", err)
	case error:
		return fmt.Errorf("Create bosh director: %s", err)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after create director: %s", err)
	}

	err = u.cloudConfigManager.Update(state)
	if err != nil {
		return fmt.Errorf("Update cloud config: %s", err)
	}

	return nil
}

func (u Up) ParseArgs(args []string, state storage.State) (UpConfig, error) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return UpConfig{}, err //not tested
	}

	prevOpsFilePath := filepath.Join(tempDir, "user-ops-file")
	err = ioutil.WriteFile(prevOpsFilePath, []byte(state.BOSH.UserOpsFile), os.ModePerm)
	if err != nil {
		return UpConfig{}, err //not tested
	}

	var config UpConfig
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
