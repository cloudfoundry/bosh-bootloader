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
	cfDNS                string
	concourseLB          string
}

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()

	if state.IAAS == "azure" && state.Azure.DisablePublicIP == "true" {
		tmpls.output = strings.Replace(tmpls.output, "azurerm_public_ip.bosh.ip_address", "cidrhost(var.internal_cidr, 5)", -1)
	}
	if state.IAAS == "azure" && state.Azure.CIDR != "" {
		tmpls.vars = strings.Replace(tmpls.vars, "10.0.0.0/16", state.Azure.CIDR, -1)
	}
	if state.IAAS == "azure" && state.Azure.CIDR != "" {
		tmpls.output = strings.Replace(tmpls.output, "${cidrsubnet(var.network_cidr, 8, 0)}", state.Azure.CIDR, -1)
	}

	template := strings.Join([]string{tmpls.vars, tmpls.resourceGroup, tmpls.network, tmpls.storage, tmpls.networkSecurityGroup, tmpls.output, tmpls.tls}, "\n")

	switch state.LB.Type {
	case "cf":
		template = strings.Join([]string{template, tmpls.cfLB}, "\n")

		if state.LB.Domain != "" {
			template = strings.Join([]string{template, tmpls.cfDNS}, "\n")
		}
	case "concourse":
		template = strings.Join([]string{template, tmpls.concourseLB}, "\n")
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
	tmpls.cfDNS = string(MustAsset("templates/cf_dns.tf"))
	tmpls.concourseLB = string(MustAsset("templates/concourse_lb.tf"))

	return tmpls
}
