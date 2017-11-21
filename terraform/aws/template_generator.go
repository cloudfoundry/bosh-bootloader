package aws

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct{}

type templates struct {
	base           string
	lbSubnet       string
	cfLB           string
	cfDNS          string
	concourseLB    string
	sslCertificate string
	isoSeg         string
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (tg TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()
	template := tmpls.base

	switch state.LB.Type {
	case "concourse":
		template = strings.Join([]string{template, tmpls.lbSubnet, tmpls.concourseLB, tmpls.sslCertificate}, "\n")
	case "cf":
		template = strings.Join([]string{template, tmpls.lbSubnet, tmpls.cfLB, tmpls.sslCertificate, tmpls.isoSeg}, "\n")

		if state.LB.Domain != "" {
			template = strings.Join([]string{template, tmpls.cfDNS}, "\n")
		}
	}

	return template
}

func readTemplates() templates {
	tmpls := templates{}
	tmpls.base = string(MustAsset("templates/base.tf"))
	tmpls.lbSubnet = string(MustAsset("templates/lb_subnet.tf"))
	tmpls.concourseLB = string(MustAsset("templates/concourse_lb.tf"))
	tmpls.sslCertificate = string(MustAsset("templates/ssl_certificate.tf"))
	tmpls.cfLB = string(MustAsset("templates/cf_lb.tf"))
	tmpls.cfDNS = string(MustAsset("templates/cf_dns.tf"))
	tmpls.isoSeg = string(MustAsset("templates/iso_segments.tf"))

	return tmpls
}
