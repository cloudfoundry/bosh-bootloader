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
	boshClientProvider        boshClientProvider
	stateStore                stateStore
}

type deleteLBsConfig struct {
	skipIfMissing bool
}

func NewAWSDeleteLBs(credentialValidator credentialValidator, availabilityZoneRetriever availabilityZoneRetriever,
	certificateManager certificateManager, infrastructureManager infrastructureManager, logger logger,
	cloudConfigManager cloudConfigManager,
	boshClientProvider boshClientProvider, stateStore stateStore,
) AWSDeleteLBs {
	return AWSDeleteLBs{
		credentialValidator:       credentialValidator,
		availabilityZoneRetriever: availabilityZoneRetriever,
		certificateManager:        certificateManager,
		infrastructureManager:     infrastructureManager,
		logger:                    logger,
		cloudConfigManager:        cloudConfigManager,
		boshClientProvider:        boshClientProvider,
		stateStore:                stateStore,
	}
}

func (c AWSDeleteLBs) Execute(state storage.State) error {
	err := c.credentialValidator.ValidateAWS()
	if err != nil {
		return err
	}

	if err := checkBBLAndLB(state, c.boshClientProvider, c.infrastructureManager); err != nil {
		return err
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

	err = c.cloudConfigManager.Update(state)
	if err != nil {
		return err
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
