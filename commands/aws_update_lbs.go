package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AWSUpdateLBs struct {
	awsCreateLBs         awsCreateLBs
	credentialValidator  credentialValidator
	environmentValidator environmentValidator
}

func NewAWSUpdateLBs(awsCreateLBs awsCreateLBs, credentialValidator credentialValidator,
	environmentValidator environmentValidator) AWSUpdateLBs {

	return AWSUpdateLBs{
		credentialValidator:  credentialValidator,
		environmentValidator: environmentValidator,
		awsCreateLBs:         awsCreateLBs,
	}
}

func (c AWSUpdateLBs) Execute(config AWSCreateLBsConfig, state storage.State) error {
	err := c.credentialValidator.Validate()
	if err != nil {
		return err
	}

	err = c.environmentValidator.Validate(state)
	if err != nil {
		return err
	}

	if config.Domain == "" {
		config.Domain = state.LB.Domain
	}
	if config.LBType == "" {
		config.LBType = state.LB.Type
	}

	return c.awsCreateLBs.Execute(config, state)
}
