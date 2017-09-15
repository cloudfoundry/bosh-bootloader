package cloudconfig

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGeneratorWrapper struct {
	awsCloudFormationOpsGenerator OpsGenerator
	awsTerraformOpsGenerator      OpsGenerator
	gcpOpsGenerator               OpsGenerator
	azureOpsGenerator             OpsGenerator
}

func NewOpsGenerator(awsCloudFormationOpsGenerator OpsGenerator, awsTerraformOpsGenerator OpsGenerator, gcpOpsGenerator OpsGenerator, azureOpsGenerator OpsGenerator) OpsGeneratorWrapper {
	return OpsGeneratorWrapper{
		awsCloudFormationOpsGenerator: awsCloudFormationOpsGenerator,
		awsTerraformOpsGenerator:      awsTerraformOpsGenerator,
		gcpOpsGenerator:               gcpOpsGenerator,
		azureOpsGenerator:             azureOpsGenerator,
	}
}

func (o OpsGeneratorWrapper) Generate(state storage.State) (string, error) {
	switch state.IAAS {
	case "gcp":
		return o.gcpOpsGenerator.Generate(state)
	case "aws":
		if state.TFState != "" {
			return o.awsTerraformOpsGenerator.Generate(state)
		} else {
			return o.awsCloudFormationOpsGenerator.Generate(state)
		}
	case "azure":
		return o.azureOpsGenerator.Generate(state)
	default:
		return "", errors.New("invalid iaas type")
	}
}
