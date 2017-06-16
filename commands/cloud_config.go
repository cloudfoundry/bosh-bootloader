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
	contents, err := c.cloudConfigManager.Generate(state)
	if err != nil {
		return err
	}
	c.logger.Println(string(contents))
	return nil
}
