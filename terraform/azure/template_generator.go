package azure

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type templates struct {
	vars                 string
	resourceGroup        string
	network              string
	storage              string
	networkSecurityGroup string
	output               string
}

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()
	return strings.Join([]string{tmpls.vars, tmpls.resourceGroup, tmpls.network, tmpls.storage, tmpls.networkSecurityGroup, tmpls.output}, "\n")
}

func readTemplates() templates {
	tmpls := templates{}
	tmpls.vars = string(MustAsset("templates/vars_template.tf"))
	tmpls.resourceGroup = string(MustAsset("templates/resource_group_template.tf"))
	tmpls.network = string(MustAsset("templates/network_template.tf"))
	tmpls.storage = string(MustAsset("templates/storage_template.tf"))
	tmpls.networkSecurityGroup = string(MustAsset("templates/network_security_group_template.tf"))
	tmpls.output = string(MustAsset("templates/output_template.tf"))

	return tmpls
}
