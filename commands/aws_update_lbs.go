package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

type AWSUpdateLBs struct {
	awsCreateLBs awsCreateLBs
}

func NewAWSUpdateLBs(awsCreateLBs awsCreateLBs) AWSUpdateLBs {
	return AWSUpdateLBs{
		awsCreateLBs: awsCreateLBs,
	}
}

func (c AWSUpdateLBs) Execute(config AWSCreateLBsConfig, state storage.State) error {
	if config.Domain == "" {
		config.Domain = state.LB.Domain
	}

	if config.LBType == "" {
		config.LBType = state.LB.Type
	}

	return c.awsCreateLBs.Execute(config, state)
}
