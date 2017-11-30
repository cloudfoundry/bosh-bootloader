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
	tls                  string
	cfLB                 string
}

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()
	template := strings.Join([]string{tmpls.vars, tmpls.resourceGroup, tmpls.network, tmpls.storage, tmpls.networkSecurityGroup, tmpls.output, tmpls.tls}, "\n")
	if state.LB.Type == "cf" {
		template = strings.Join([]string{template, tmpls.cfLB}, "\n")
	}

	return template
}

func readTemplates() templates {
	tmpls := templates{}
	tmpls.vars = string(MustAsset("templates/vars.tf"))
	tmpls.resourceGroup = string(MustAsset("templates/resource_group.tf"))
	tmpls.network = string(MustAsset("templates/network.tf"))
	tmpls.storage = string(MustAsset("templates/storage.tf"))
	tmpls.networkSecurityGroup = string(MustAsset("templates/network_security_group.tf"))
	tmpls.output = string(MustAsset("templates/output.tf"))
	tmpls.tls = string(MustAsset("templates/tls.tf"))
	tmpls.cfLB = string(MustAsset("templates/cf_lb.tf"))

	return tmpls
}
