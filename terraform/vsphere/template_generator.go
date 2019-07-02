package vsphere

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	return fmt.Sprint(`
variable "env_id" {}
variable "director_internal_ip" {}
variable "internal_gw" {}
variable "jumpbox_ip" {}
variable "network_name" {}
variable "vcenter_cluster" {}
variable "vsphere_subnet_cidr" {}
variable "vcenter_user" {}
variable "vcenter_password" {}
variable "vcenter_ip" {}
variable "vcenter_dc" {}
variable "vcenter_rp" {}
variable "vcenter_ds" {}
variable "vcenter_templates" {}
variable "vcenter_vms" {}
variable "vcenter_disks" {}

output "internal_cidr" { value = "${var.vsphere_subnet_cidr}" }
output "internal_gw" { value = "${var.internal_gw}" }
output "network_name" { value = "${var.network_name}" }
output "vcenter_cluster" { value = "${var.vcenter_cluster}" }
output "jumpbox_url" { value = "${var.jumpbox_ip}:22" }
output "external_ip" { value = "${var.jumpbox_ip}" }
output "jumpbox__internal_ip" { value = "${var.jumpbox_ip}" }
output "director__internal_ip" { value = "${var.director_internal_ip}" }
output "director_name" { value = "bosh-${var.env_id}" }
output "vcenter_disks" { value = "${var.vcenter_disks}" }
output "vcenter_vms" { value = "${var.vcenter_vms}" }
output "vcenter_templates" { value = "${var.vcenter_templates}" }
output "vcenter_ip" { value = "${var.vcenter_ip}" }
output "vcenter_dc" { value = "${var.vcenter_dc}" }
output "vcenter_rp" { value = "${var.vcenter_rp}" }
output "vcenter_ds" { value = "${var.vcenter_ds}" }
`)
}
