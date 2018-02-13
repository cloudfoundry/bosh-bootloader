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
variable "external_ip" {}
variable "auth_url" {}
variable "az" {}
variable "default_key_name" {}
variable "default_security_groups" {}
variable "net_id" {}
variable "openstack_project" {}
variable "openstack_domain" {}
variable "region" {}
variable "env_id" {}

output "internal_cidr" { value = "${var.internal_cidr}" }
output "external_ip" { value = "${var.external_ip}" }
output "auth_url" { value = "${var.auth_url}" }
output "az" { value = "${var.az}" }
output "default_key_name" { value = "${var.default_key_name}" }
output "default_security_group" { value = "${var.default_security_group}" }
output "net_id" { value = "${var.net_id}" }
output "openstack_project" { value = "${var.openstack_project}" }
output "region" { value = "${var.region}" }
output "env_id" { value = "${var.env_id}" }
`)
}
