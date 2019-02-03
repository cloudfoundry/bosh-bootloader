package openstack

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type templates struct {
	networkOutputs   string
	networkVars      string
	network          string
	providerVars     string
	provider         string
	resourcesOutputs string
	resourcesVars    string
	resources        string
}

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()
	template := strings.Join([]string{tmpls.networkOutputs, tmpls.networkVars, tmpls.network, tmpls.providerVars, tmpls.provider, tmpls.resourcesOutputs, tmpls.resourcesVars, tmpls.resources}, "\n")
	return template
}

func readTemplates() templates {
	tmpls := templates{}
	tmpls.networkOutputs = string(MustAsset("templates/network-outputs.tf"))
	tmpls.networkVars = string(MustAsset("templates/network-vars.tf"))
	tmpls.network = string(MustAsset("templates/network.tf"))
	tmpls.providerVars = string(MustAsset("templates/provider-vars.tf"))
	tmpls.provider = string(MustAsset("templates/provider.tf"))
	tmpls.resourcesOutputs = string(MustAsset("templates/resources-outputs.tf"))
	tmpls.resourcesVars = string(MustAsset("templates/resources-vars.tf"))
	tmpls.resources = string(MustAsset("templates/resources.tf"))

	return tmpls
}
