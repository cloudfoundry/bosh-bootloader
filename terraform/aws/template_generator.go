package aws

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct{}

type templates struct {
	base           string
	iam            string
	lbSubnet       string
	cfLB           string
	cfDNS          string
	concourseLB    string
	sslCertificate string
	isoSeg         string
	vpc            string
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (tg TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()
	template := strings.Join([]string{tmpls.base, tmpls.iam, tmpls.vpc}, "\n")

	switch state.LB.Type {
	case "concourse":
		template = strings.Join([]string{template, tmpls.lbSubnet, tmpls.concourseLB}, "\n")
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
	tmpls.iam = string(MustAsset("templates/iam.tf"))
	tmpls.lbSubnet = string(MustAsset("templates/lb_subnet.tf"))
	tmpls.concourseLB = string(MustAsset("templates/concourse_lb.tf"))
	tmpls.sslCertificate = string(MustAsset("templates/ssl_certificate.tf"))
	tmpls.cfLB = string(MustAsset("templates/cf_lb.tf"))
	tmpls.cfDNS = string(MustAsset("templates/cf_dns.tf"))
	tmpls.isoSeg = string(MustAsset("templates/iso_segments.tf"))
	tmpls.vpc = string(MustAsset("templates/vpc.tf"))

	return tmpls
}
