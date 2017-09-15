package aws

import (
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGenerator struct {
	awsCloudFormationOpsGenerator cloudconfig.OpsGenerator
	awsTerraformOpsGenerator      cloudconfig.OpsGenerator
}

func NewOpsGenerator(awsCloudFormationOpsGenerator cloudconfig.OpsGenerator, awsTerraformOpsGenerator cloudconfig.OpsGenerator) OpsGenerator {
	return OpsGenerator{
		awsCloudFormationOpsGenerator: awsCloudFormationOpsGenerator,
		awsTerraformOpsGenerator:      awsTerraformOpsGenerator,
	}
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	if state.TFState != "" {
		return o.awsTerraformOpsGenerator.Generate(state)
	} else {
		return o.awsCloudFormationOpsGenerator.Generate(state)
	}
}
