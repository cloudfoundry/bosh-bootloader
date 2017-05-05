package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

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
	terraformManager          terraformManager
}

type deleteLBsConfig struct {
	skipIfMissing bool
}

func NewAWSDeleteLBs(credentialValidator credentialValidator, availabilityZoneRetriever availabilityZoneRetriever,
	certificateManager certificateManager, infrastructureManager infrastructureManager, logger logger,
	cloudConfigManager cloudConfigManager, stateStore stateStore,
	environmentValidator environmentValidator, terraformManager terraformManager,
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
		terraformManager:          terraformManager,
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

	if state.TFState != "" {
		if !lbExists(state.LB.Type) {
			return LBNotFound
		}

		state.LB.Type = ""
		state.LB.Cert = ""
		state.LB.Key = ""
	} else {
		if !lbExists(state.Stack.LBType) {
			return LBNotFound
		}

		_, err = c.infrastructureManager.Describe(state.Stack.Name)
		if err != nil {
			return err
		}

		state.Stack.LBType = "none"
	}

	if !state.NoDirector {
		err = c.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	if state.TFState != "" {
		state, err = c.terraformManager.Apply(state)
		if err != nil {
			return handleTerraformError(err, c.stateStore)
		}
	} else {
		azs, err := c.availabilityZoneRetriever.Retrieve(state.AWS.Region)
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
	}

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
