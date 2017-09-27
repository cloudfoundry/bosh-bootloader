package cloudconfig

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGeneratorWrapper struct {
	awsOpsGenerator   OpsGenerator
	gcpOpsGenerator   OpsGenerator
	azureOpsGenerator OpsGenerator
}

func NewOpsGenerator(awsOpsGenerator OpsGenerator, gcpOpsGenerator OpsGenerator, azureOpsGenerator OpsGenerator) OpsGeneratorWrapper {
	return OpsGeneratorWrapper{
		awsOpsGenerator:   awsOpsGenerator,
		gcpOpsGenerator:   gcpOpsGenerator,
		azureOpsGenerator: azureOpsGenerator,
	}
}

func (o OpsGeneratorWrapper) Generate(state storage.State) (string, error) {
	switch state.IAAS {
	case "gcp":
		return o.gcpOpsGenerator.Generate(state)
	case "aws":
		return o.awsOpsGenerator.Generate(state)
	case "azure":
		return o.azureOpsGenerator.Generate(state)
	default:
		return "", errors.New("invalid iaas type")
	}
}
