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

type Plan struct {
	boshManager        boshManager
	cloudConfigManager cloudConfigManager
	stateStore         stateStore
	envIDManager       envIDManager
	terraformManager   terraformManager
	lbArgsHandler      lbArgsHandler
}

type PlanConfig struct {
	Name       string
	OpsFile    string
	NoDirector bool
	LB         CreateLBsConfig
}

func NewPlan(boshManager boshManager, cloudConfigManager cloudConfigManager,
	stateStore stateStore, envIDManager envIDManager, terraformManager terraformManager,
	lbArgsHandler lbArgsHandler) Plan {
	return Plan{
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		stateStore:         stateStore,
		envIDManager:       envIDManager,
		terraformManager:   terraformManager,
		lbArgsHandler:      lbArgsHandler,
	}
}

func (p Plan) CheckFastFails(args []string, state storage.State) error {
	config, err := p.ParseArgs(args, state)
	if err != nil {
		return err
	}

	if !config.NoDirector && !state.NoDirector {
		if err := fastFailBOSHVersion(p.boshManager); err != nil {
			return err
		}
	}

	if err := p.terraformManager.ValidateVersion(); err != nil {
		return fmt.Errorf("Terraform manager validate version: %s", err)
	}

	if state.EnvID != "" && config.Name != "" && config.Name != state.EnvID {
		return fmt.Errorf("The director name cannot be changed for an existing environment. Current name is %s.", state.EnvID)
	}

	return nil
}

func (p Plan) ParseArgs(args []string, state storage.State) (PlanConfig, error) {
	opsFileDir, err := p.stateStore.GetBblDir()
	if err != nil {
		return PlanConfig{}, err //not tested
	}

	prevOpsFilePath := filepath.Join(opsFileDir, "previous-user-ops-file.yml")
	err = ioutil.WriteFile(prevOpsFilePath, []byte(state.BOSH.UserOpsFile), os.ModePerm)
	if err != nil {
		return PlanConfig{}, err //not tested
	}

	var config PlanConfig
	planFlags := flags.New("up")
	planFlags.String(&config.Name, "name", "")
	planFlags.String(&config.OpsFile, "ops-file", prevOpsFilePath)
	planFlags.Bool(&config.NoDirector, "", "no-director", state.NoDirector)
	planFlags.String(&config.LB.LBType, "lb-type", "")
	planFlags.String(&config.LB.CertPath, "lb-cert", "")
	planFlags.String(&config.LB.KeyPath, "lb-key", "")
	planFlags.String(&config.LB.Domain, "lb-domain", "")
	if state.IAAS == "aws" {
		planFlags.String(&config.LB.ChainPath, "lb-chain", "")
	}

	err = planFlags.Parse(args)
	if err != nil {
		return PlanConfig{}, err
	}

	if (config.LB != CreateLBsConfig{}) {
		_, err = p.lbArgsHandler.GetLBState(state.IAAS, config.LB)
		if err != nil {
			return PlanConfig{}, err
		}
	}

	return config, nil
}

func (p Plan) Execute(args []string, state storage.State) error {
	config, err := p.ParseArgs(args, state)
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

	state, err = p.envIDManager.Sync(state, config.Name)
	if err != nil {
		return fmt.Errorf("Env id manager sync: %s", err)
	}

	err = p.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state: %s", err)
	}

	if err := p.terraformManager.Init(state); err != nil {
		return fmt.Errorf("Terraform manager init: %s", err)
	}

	if state.NoDirector {
		return nil
	}

	if err := p.boshManager.InitializeJumpbox(state); err != nil {
		return fmt.Errorf("Bosh manager initialize jumpbox: %s", err)
	}

	state.BOSH.UserOpsFile = string(opsFileContents)
	if err := p.boshManager.InitializeDirector(state); err != nil {
		return fmt.Errorf("Bosh manager initialize director: %s", err)
	}

	if err := p.cloudConfigManager.Initialize(state); err != nil {
		return fmt.Errorf("Cloud config manager initialize: %s", err)
	}

	return nil
}

func (p Plan) IsInitialized(state storage.State) bool {
	return state.BBLVersion != "" && state.BBLVersion >= "5.2.0"
}
