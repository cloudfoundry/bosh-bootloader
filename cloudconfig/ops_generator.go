package cloudconfig

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGenerator struct {
	awsOpsGenerator opsGenerator
	gcpOpsGenerator opsGenerator
}

func NewOpsGenerator(awsOpsGenerator opsGenerator, gcpOpsGenerator opsGenerator) OpsGenerator {
	return OpsGenerator{
		awsOpsGenerator: awsOpsGenerator,
		gcpOpsGenerator: gcpOpsGenerator,
	}
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	switch state.IAAS {
	case "gcp":
		return o.gcpOpsGenerator.Generate(state)
	case "aws":
		return o.awsOpsGenerator.Generate(state)
	default:
		return "", errors.New("invalid iaas type")
	}
}
