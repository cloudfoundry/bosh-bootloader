package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TemplateGenerator struct {
	gcpTemplateGenerator templateGenerator
	awsTemplateGenerator templateGenerator
}

func NewTemplateGenerator(gcpTemplateGenerator templateGenerator, awsTemplateGenerator templateGenerator) TemplateGenerator {
	return TemplateGenerator{
		gcpTemplateGenerator: gcpTemplateGenerator,
		awsTemplateGenerator: awsTemplateGenerator,
	}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	switch state.IAAS {
	case "gcp":
		return t.gcpTemplateGenerator.Generate(state)
	case "aws":
		return t.awsTemplateGenerator.Generate(state)
	default:
		return ""
	}
}
