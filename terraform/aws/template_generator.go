package aws

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var AMIs = `{
	"us-east-1":      "ami-68115b02",
	"us-east-2":      "ami-6893b20d",
	"us-west-1":      "ami-ef1a718f",
	"us-west-2":      "ami-77a4b816",
	"eu-west-1":      "ami-c0993ab3",
	"eu-central-1":   "ami-0b322e67",
	"ap-southeast-1": "ami-e2fc3f81",
	"ap-southeast-2": "ami-e3217a80",
	"ap-northeast-1": "ami-f885ae96",
	"ap-northeast-2": "ami-4118d72f",
	"sa-east-1":      "ami-8631b5ea"
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

	var ami map[string]string

	err := json.Unmarshal([]byte(AMIs), &ami)
	if err != nil {
		panic(err)
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
			AWSNATAMIs:                   ami,
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
			AWSNATAMIs:                   ami,
		}
	}

	if state.LB.Cert == "" || state.LB.Key == "" {
		templateData.IgnoreSSLCertificateProperties = `ignore_changes = ["certificate_body", "certificate_chain", "private_key"]`
	}

	tmpl := template.New("descriptions")
	tmpl, err = tmpl.Parse(t)
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
