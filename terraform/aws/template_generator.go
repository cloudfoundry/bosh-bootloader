package aws

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var AMIs = `{
  "ap-northeast-1":  "ami-10dfc877",
  "ap-northeast-2":  "ami-1a1bc474",
  "ap-south-1": "ami-74c1861b",
  "ap-southeast-1":  "ami-36af2055",
  "ap-southeast-2":  "ami-1e91817d",
  "eu-central-1":  "ami-9ebe18f1",
  "eu-west-1":  "ami-3a849f5c",
  "eu-west-2":  "ami-21120445",
  "us-east-1":  "ami-d4c5efc2",
  "us-east-2":  "ami-f27b5a97",
  "us-gov-west-1":  "ami-c39610a2",
  "us-west-1":  "ami-b87f53d8",
  "us-west-2":  "ami-8bfce8f2"
}`

type TemplateGenerator struct{}

type TemplateData struct {
	NATDescription                 string
	InternalDescription            string
	BOSHDescription                string
	ConcourseDescription           string
	ConcourseInternalDescription   string
	SSHLBDescription               string
	SSHLBInternalDescription       string
	RouterDescription              string
	RouterInternalDescription      string
	TCPLBDescription               string
	TCPLBInternalDescription       string
	SSLCertificateNameProperty     string
	IgnoreSSLCertificateProperties string
	AWSNATAMIs                     map[string]string
}

type templates struct {
	base           string
	lbSubnet       string
	cfLB           string
	cfDNS          string
	concourseLB    string
	sslCertificate string
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (tg TemplateGenerator) Generate(state storage.State) string {
	tmpls := readTemplates()
	tmpl := tmpls.base

	switch state.LB.Type {
	case "concourse":
		tmpl = strings.Join([]string{tmpl, tmpls.lbSubnet, tmpls.concourseLB, tmpls.sslCertificate}, "\n")
	case "cf":
		tmpl = strings.Join([]string{tmpl, tmpls.lbSubnet, tmpls.cfLB, tmpls.sslCertificate}, "\n")

		if state.LB.Domain != "" {
			tmpl = strings.Join([]string{tmpl, tmpls.cfDNS}, "\n")
		}
	}

	var ami map[string]string
	err := json.Unmarshal([]byte(AMIs), &ami)
	if err != nil {
		panic(err)
	}

	templateData := TemplateData{
		AWSNATAMIs:                   ami,
		BOSHDescription:              "Bosh",
		ConcourseDescription:         "Concourse",
		ConcourseInternalDescription: "Concourse Internal",
		InternalDescription:          "Internal",
		NATDescription:               "NAT",
		RouterDescription:            "CF Router",
		RouterInternalDescription:    "CF Router Internal",
		SSHLBDescription:             "CF SSH",
		SSHLBInternalDescription:     "CF SSH Internal",
		SSLCertificateNameProperty:   `name_prefix       = "${var.ssl_certificate_name_prefix}"`,
		TCPLBDescription:             "CF TCP",
		TCPLBInternalDescription:     "CF TCP Internal",
	}

	t := template.New("descriptions")
	t, err = t.Parse(tmpl)
	if err != nil {
		panic(err)
	}

	finalTemplate := bytes.Buffer{}

	err = t.Execute(&finalTemplate, templateData)
	if err != nil {
		panic(err)
	}

	return finalTemplate.String()
}

func readTemplates() templates {
	tmpls := templates{}
	tmpls.base = string(MustAsset("templates/base.tf"))
	tmpls.lbSubnet = string(MustAsset("templates/lb_subnet.tf"))
	tmpls.concourseLB = string(MustAsset("templates/concourse_lb.tf"))
	tmpls.sslCertificate = string(MustAsset("templates/ssl_certificate.tf"))
	tmpls.cfLB = string(MustAsset("templates/cf_lb.tf"))
	tmpls.cfDNS = string(MustAsset("templates/cf_dns.tf"))

	return tmpls
}
