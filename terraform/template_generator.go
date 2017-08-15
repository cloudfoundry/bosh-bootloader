package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TemplateGenerator struct {
	gcpTemplateGenerator   templateGenerator
	awsTemplateGenerator   templateGenerator
	azureTemplateGenerator templateGenerator
}

func NewTemplateGenerator(gcpTemplateGenerator templateGenerator, awsTemplateGenerator templateGenerator, azureTemplateGenerator templateGenerator) TemplateGenerator {
	return TemplateGenerator{
		gcpTemplateGenerator:   gcpTemplateGenerator,
		awsTemplateGenerator:   awsTemplateGenerator,
		azureTemplateGenerator: azureTemplateGenerator,
	}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	switch state.IAAS {
	case "gcp":
		return t.gcpTemplateGenerator.Generate(state)
	case "aws":
		return t.awsTemplateGenerator.Generate(state)
	case "azure":
		return t.azureTemplateGenerator.Generate(state)
	default:
		return ""
	}
}
