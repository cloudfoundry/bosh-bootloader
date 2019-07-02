package openstack

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/gobuffalo/packr/v2"
)

const templatesPath = "./templates"

type templates struct {
	providerVars     string
	provider         string
	resourcesOutputs string
	resourcesVars    string
	resources        string
}

type TemplateGenerator struct {
	box *packr.Box
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{
		box: packr.New("openstack-templates", templatesPath),
	}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	tmpls := t.readTemplates()
	template := strings.Join([]string{tmpls.providerVars, tmpls.provider, tmpls.resourcesOutputs, tmpls.resourcesVars, tmpls.resources}, "\n")
	return template
}

func (t TemplateGenerator) readTemplates() templates {
	listings := map[string]string{
		"provider-vars.tf":     "",
		"provider.tf":          "",
		"resources-outputs.tf": "",
		"resources-vars.tf":    "",
		"resources.tf":         "",
	}

	var errors []error
	for item := range listings {
		content, err := t.box.FindString(item)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		listings[item] = content
	}

	if errors != nil {
		panic(errors)
	}

	return templates{
		providerVars:     listings["provider-vars.tf"],
		provider:         listings["provider.tf"],
		resourcesOutputs: listings["resources-outputs.tf"],
		resourcesVars:    listings["resources-vars.tf"],
		resources:        listings["resources.tf"],
	}
}
