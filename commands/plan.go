package commands

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type patchDetector interface {
	Find() error
}

type Plan struct {
	boshManager          boshManager
	cloudConfigManager   cloudConfigManager
	runtimeConfigManager runtimeConfigManager
	stateStore           stateStore
	patchDetector        patchDetector
	envIDManager         envIDManager
	terraformManager     terraformManager
	lbArgsHandler        lbArgsHandler
	logger               logger
	bblVersion           string
}

type PlanConfig struct {
	Name string
	LB   storage.LB
}

func NewPlan(
	boshManager boshManager,
	cloudConfigManager cloudConfigManager,
	runtimeConfigManager runtimeConfigManager,
	stateStore stateStore,
	patchDetector patchDetector,
	envIDManager envIDManager,
	terraformManager terraformManager,
	lbArgsHandler lbArgsHandler,
	logger logger,
	bblVersion string,
) Plan {
	return Plan{
		boshManager:          boshManager,
		cloudConfigManager:   cloudConfigManager,
		runtimeConfigManager: runtimeConfigManager,
		stateStore:           stateStore,
		patchDetector:        patchDetector,
		envIDManager:         envIDManager,
		terraformManager:     terraformManager,
		lbArgsHandler:        lbArgsHandler,
		logger:               logger,
		bblVersion:           bblVersion,
	}
}

func (p Plan) CheckFastFails(args []string, state storage.State) error {
	config, err := p.ParseArgs(args, state)
	if err != nil {
		return err
	}

	if err := fastFailBOSHVersion(p.boshManager); err != nil {
		return err
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
		config PlanConfig
		lbArgs LBArgs
	)
	planFlags := flags.New("up")
	planFlags.String(&config.Name, "name", os.Getenv("BBL_ENV_NAME"))
	planFlags.String(&lbArgs.LBType, "lb-type", "")
	planFlags.String(&lbArgs.CertPath, "lb-cert", "")
	planFlags.String(&lbArgs.KeyPath, "lb-key", "")
	planFlags.String(&lbArgs.Domain, "lb-domain", "")

	err := planFlags.Parse(args)
	if err != nil {
		return PlanConfig{}, err
	}

	if (lbArgs != LBArgs{}) {
		lbState, err := p.lbArgsHandler.GetLBState(state.IAAS, lbArgs)
		if err != nil {
			return PlanConfig{}, err
		}
		config.LB = lbState
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
	state.BBLVersion = p.bblVersion
	state.LB = config.LB

	var err error
	state, err = p.envIDManager.Sync(state, config.Name)
	if err != nil {
		return storage.State{}, fmt.Errorf("Env id manager sync: %s", err)
	}

	err = p.stateStore.Set(state)
	if err != nil {
		return storage.State{}, fmt.Errorf("Save state: %s", err)
	}

	if err := p.terraformManager.Setup(state); err != nil {
		return storage.State{}, fmt.Errorf("Terraform manager init: %s", err)
	}

	if err := p.boshManager.InitializeJumpbox(state); err != nil {
		return storage.State{}, fmt.Errorf("Bosh manager initialize jumpbox: %s", err)
	}

	if err := p.boshManager.InitializeDirector(state); err != nil {
		return storage.State{}, fmt.Errorf("Bosh manager initialize director: %s", err)
	}

	if err := p.cloudConfigManager.Initialize(state); err != nil {
		return storage.State{}, fmt.Errorf("Cloud config manager initialize: %s", err)
	}

	if err := p.runtimeConfigManager.Initialize(state); err != nil {
		return storage.State{}, fmt.Errorf("Runtime config manager initialize: %s", err)
	}

	if err := p.patchDetector.Find(); err != nil {
		p.logger.Printf("Failed to detect patch files: %s\n", err)
	}

	return state, nil
}

func (p Plan) IsInitialized(state storage.State) bool {
	// If it is older than bbl v5.4.0 with schema 13, we want to re-initialize.
	return state.Version >= 13
}
