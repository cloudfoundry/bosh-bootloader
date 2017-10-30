package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const (
	CloudConfigCommand = "cloud-config"
)

type CloudConfig struct {
	logger             logger
	stateValidator     stateValidator
	cloudConfigManager cloudConfigManager
}

func NewCloudConfig(logger logger, stateValidator stateValidator, cloudConfigManager cloudConfigManager) CloudConfig {
	return CloudConfig{
		logger:             logger,
		stateValidator:     stateValidator,
		cloudConfigManager: cloudConfigManager,
	}
}

func (c CloudConfig) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := c.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (c CloudConfig) Execute(args []string, state storage.State) error {
	if !c.cloudConfigManager.IsPresentCloudConfig() {
		err := c.cloudConfigManager.Initialize(state)
		if err != nil {
			return err
		}
	}
	if !c.cloudConfigManager.IsPresentCloudConfigVars() {
		err := c.cloudConfigManager.GenerateVars(state)
		if err != nil {
			return err
		}
	}
	contents, err := c.cloudConfigManager.Interpolate()
	if err != nil {
		return err
	}
	c.logger.Println(string(contents))
	return nil
}
