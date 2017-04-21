package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	DeleteLBsCommand = "delete-lbs"
)

type AWSDeleteLBs struct {
	credentialValidator       credentialValidator
	availabilityZoneRetriever availabilityZoneRetriever
	certificateManager        certificateManager
	infrastructureManager     infrastructureManager
	logger                    logger
	cloudConfigManager        cloudConfigManager
	stateStore                stateStore
	environmentValidator      environmentValidator
}

type deleteLBsConfig struct {
	skipIfMissing bool
}

func NewAWSDeleteLBs(credentialValidator credentialValidator, availabilityZoneRetriever availabilityZoneRetriever,
	certificateManager certificateManager, infrastructureManager infrastructureManager, logger logger,
	cloudConfigManager cloudConfigManager, stateStore stateStore,
	environmentValidator environmentValidator,
) AWSDeleteLBs {
	return AWSDeleteLBs{
		credentialValidator:       credentialValidator,
		availabilityZoneRetriever: availabilityZoneRetriever,
		certificateManager:        certificateManager,
		infrastructureManager:     infrastructureManager,
		logger:                    logger,
		cloudConfigManager:        cloudConfigManager,
		stateStore:                stateStore,
		environmentValidator:      environmentValidator,
	}
}

func (c AWSDeleteLBs) Execute(state storage.State) error {
	err := c.credentialValidator.Validate()
	if err != nil {
		return err
	}

	err = c.environmentValidator.Validate(state)
	if err != nil {
		return err
	}

	if !lbExists(state.Stack.LBType) {
		return LBNotFound
	}

	azs, err := c.availabilityZoneRetriever.Retrieve(state.AWS.Region)
	if err != nil {
		return err
	}

	_, err = c.infrastructureManager.Describe(state.Stack.Name)
	if err != nil {
		return err
	}

	state.Stack.LBType = "none"

	if !state.NoDirector {
		err = c.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	_, err = c.infrastructureManager.Update(state.KeyPair.Name, azs, state.Stack.Name, state.Stack.BOSHAZ, "", "", state.EnvID)
	if err != nil {
		return err
	}

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	c.logger.Step("deleting certificate")
	err = c.certificateManager.Delete(state.Stack.CertificateName)
	if err != nil {
		return err
	}

	state.Stack.CertificateName = ""

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
