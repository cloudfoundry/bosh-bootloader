package aws

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct{}

type TemplateData struct {
	NATDescription               string
	InternalDescription          string
	BOSHDescription              string
	ConcourseDescription         string
	ConcourseInternalDescription string
	SSHLBDescription             string
	SSHLBInternalDescription     string
	RouterDescription            string
	RouterInternalDescription    string
	TCPLBDescription             string
	TCPLBInternalDescription     string
	SSLCertificateNameProperty   string
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (tg TemplateGenerator) Generate(state storage.State) string {
	t := BaseTemplate

	switch state.LB.Type {
	case "concourse":
		t = strings.Join([]string{t, LBSubnetTemplate, ConcourseLBTemplate, SSLCertificateTemplate}, "\n")
	case "cf":
		t = strings.Join([]string{t, LBSubnetTemplate, CFLBTemplate, SSLCertificateTemplate}, "\n")

		if state.LB.Domain != "" {
			t = strings.Join([]string{t, CFDNSTemplate}, "\n")
		}
	}

	var templateData TemplateData
	if state.MigratedFromCloudFormation {
		templateData = TemplateData{
			NATDescription:               "NAT",
			InternalDescription:          "Internal",
			BOSHDescription:              "BOSH",
			ConcourseDescription:         "Concourse",
			ConcourseInternalDescription: "Concourse Internal",
			SSHLBDescription:             "CFSSHProxy",
			SSHLBInternalDescription:     "CFSSHProxyInternal",
			RouterDescription:            "Router",
			RouterInternalDescription:    "CFRouterInternal",
			TCPLBDescription:             "CF TCP",
			TCPLBInternalDescription:     "CF TCP Internal",
			SSLCertificateNameProperty:   `name              = "${var.ssl_certificate_name}"`,
		}
	} else {
		templateData = TemplateData{
			NATDescription:               "NAT",
			InternalDescription:          "Internal",
			BOSHDescription:              "Bosh",
			ConcourseDescription:         "Concourse",
			ConcourseInternalDescription: "Concourse Internal",
			SSHLBDescription:             "CF SSH",
			SSHLBInternalDescription:     "CF SSH Internal",
			RouterDescription:            "CF Router",
			RouterInternalDescription:    "CF Router Internal",
			TCPLBDescription:             "CF TCP",
			TCPLBInternalDescription:     "CF TCP Internal",
			SSLCertificateNameProperty:   `name_prefix       = "${var.ssl_certificate_name_prefix}"`,
		}
	}

	tmpl := template.New("descriptions")
	tmpl, err := tmpl.Parse(t)
	if err != nil {
		panic(err)
	}

	finalTemplate := bytes.Buffer{}

	err = tmpl.Execute(&finalTemplate, templateData)
	if err != nil {
		panic(err)
	}

	return finalTemplate.String()
}
