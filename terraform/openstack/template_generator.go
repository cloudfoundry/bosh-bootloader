package openstack

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	return fmt.Sprintf(`
variable "internal_cidr" {}
variable "internal_gw" {}
variable "director_ip" {}
variable "external_ip" {}
variable "auth_url" {}
variable "az" {}
variable "default_key_name" {}
variable "default_security_group" {}
variable "net_id" {}
variable "openstack_project" {}
variable "openstack_domain" {}
variable "region" {}
variable "env_id" {}

output "internal_cidr" { value = "${var.internal_cidr}" }
output "internal_gw" { value = "${var.internal_gw}" }
output "external_ip" { value = "${var.external_ip}" }
output "jumpbox__internal_ip" { value = "${var.external_ip}" }
output "director__internal_ip" { value = "${var.director_ip}" }
output "auth_url" { value = "${var.auth_url}" }
output "az" { value = "${var.az}" }
output "default_key_name" { value = "${var.default_key_name}" }
output "default_security_groups" { value = "[${var.default_security_group}]" }
output "net_id" { value = "${var.net_id}" }
output "openstack_project" { value = "${var.openstack_project}" }
output "region" { value = "${var.region}" }
output "env_id" { value = "${var.env_id}" }
`)
}
