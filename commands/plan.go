package commands

import (
	"errors"
	"fmt"
	"io/ioutil"

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
	logger             logger
}

type PlanConfig struct {
	Name       string
	OpsFile    string
	NoDirector bool
	LB         storage.LB
}

func NewPlan(boshManager boshManager,
	cloudConfigManager cloudConfigManager,
	stateStore stateStore,
	envIDManager envIDManager,
	terraformManager terraformManager,
	lbArgsHandler lbArgsHandler,
	logger logger,
) Plan {
	return Plan{
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		stateStore:         stateStore,
		envIDManager:       envIDManager,
		terraformManager:   terraformManager,
		lbArgsHandler:      lbArgsHandler,
		logger:             logger,
	}
}

func (p Plan) CheckFastFails(args []string, state storage.State) error {
	config, err := p.ParseArgs(args, state)
	if err != nil {
		return err
	}

	if config.NoDirector {
		p.logger.Println(`Deprecation warning: --no-director has been deprecated and will be removed in bbl v6.0.0. Use "bbl plan" to perform advanced configuration of the BOSH director.`)
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
	var (
		config      PlanConfig
		lbConfig    CreateLBsConfig
		opsFilePath string
	)
	planFlags := flags.New("up")
	planFlags.String(&config.Name, "name", "")
	planFlags.String(&opsFilePath, "ops-file", "")
	planFlags.Bool(&config.NoDirector, "", "no-director", state.NoDirector)
	planFlags.String(&lbConfig.LBType, "lb-type", "")
	planFlags.String(&lbConfig.CertPath, "lb-cert", "")
	planFlags.String(&lbConfig.KeyPath, "lb-key", "")
	planFlags.String(&lbConfig.Domain, "lb-domain", "")
	if state.IAAS == "aws" {
		planFlags.String(&lbConfig.ChainPath, "lb-chain", "")
	}

	err := planFlags.Parse(args)
	if err != nil {
		return PlanConfig{}, err
	}

	if (lbConfig != CreateLBsConfig{}) {
		lbState, err := p.lbArgsHandler.GetLBState(state.IAAS, lbConfig)
		if err != nil {
			return PlanConfig{}, err
		}
		config.LB = lbState
	}

	if opsFilePath != "" {
		p.logger.Println(`Deprecation warning: the --ops-file flag is now deprecated and will be removed in bbl v6.0.0. Use "bbl plan" and modify create-director.sh in your state directory to supply operations files for bosh-deployment.`)

		opsFileContents, err := ioutil.ReadFile(opsFilePath)
		if err != nil {
			return PlanConfig{}, fmt.Errorf("Reading ops-file contents: %v", err)
		}
		config.OpsFile = string(opsFileContents)
	} else {
		config.OpsFile = state.BOSH.UserOpsFile
	}

	return config, nil
}

func (p Plan) Execute(args []string, state storage.State) error {
	config, err := p.ParseArgs(args, state)
	if err != nil {
		return err
	}

	_, err = p.InitializePlan(config, state)
	return err
}

func (p Plan) InitializePlan(config PlanConfig, state storage.State) (storage.State, error) {
	if config.NoDirector {
		if !state.BOSH.IsEmpty() {
			return storage.State{}, errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}
		state.NoDirector = true
	}

	var err error

	state.LB = config.LB

	state, err = p.envIDManager.Sync(state, config.Name)
	if err != nil {
		return storage.State{}, fmt.Errorf("Env id manager sync: %s", err)
	}

	err = p.stateStore.Set(state)
	if err != nil {
		return storage.State{}, fmt.Errorf("Save state: %s", err)
	}

	if err := p.terraformManager.Init(state); err != nil {
		return storage.State{}, fmt.Errorf("Terraform manager init: %s", err)
	}

	if err := p.cloudConfigManager.Initialize(state); err != nil {
		return storage.State{}, fmt.Errorf("Cloud config manager initialize: %s", err)
	}

	if state.NoDirector {
		return state, nil
	}

	if err := p.boshManager.InitializeJumpbox(state); err != nil {
		return storage.State{}, fmt.Errorf("Bosh manager initialize jumpbox: %s", err)
	}

	state.BOSH.UserOpsFile = config.OpsFile
	if err := p.boshManager.InitializeDirector(state); err != nil {
		return storage.State{}, fmt.Errorf("Bosh manager initialize director: %s", err)
	}

	return state, nil
}

func (p Plan) IsInitialized(state storage.State) bool {
	return state.BBLVersion != "" && state.BBLVersion >= "5.2.0"
}
