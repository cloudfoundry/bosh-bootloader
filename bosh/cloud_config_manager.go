package bosh

import "gopkg.in/yaml.v2"

type CloudConfigManager struct {
	logger               logger
	cloudConfigGenerator cloudConfigGenerator
}

type cloudConfigGenerator interface {
	Generate(CloudConfigInput) (CloudConfig, error)
}

func NewCloudConfigManager(logger logger, cloudConfigGenerator cloudConfigGenerator) CloudConfigManager {
	return CloudConfigManager{
		logger:               logger,
		cloudConfigGenerator: cloudConfigGenerator,
	}
}

func (c CloudConfigManager) Update(input CloudConfigInput, boshClient Client) error {
	c.logger.Step("generating cloud config")
	cloudConfig, err := c.cloudConfigGenerator.Generate(input)
	if err != nil {
		return err
	}

	manifestYAML, err := yaml.Marshal(cloudConfig)
	if err != nil {
		return err
	}

	c.logger.Step("applying cloud config")
	if err := boshClient.UpdateCloudConfig(manifestYAML); err != nil {
		return err
	}

	return nil
}
