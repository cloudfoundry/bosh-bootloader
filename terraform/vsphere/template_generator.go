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
	return fmt.Sprintf(`
variable "bosh_director_internal_ip" {}
variable "internal_gw" {}
variable "jumpbox_ip" {}
variable "network_name" {}
variable "vcenter_cluster" {}
variable "vsphere_subnet" {}
variable "vcenter_user" {}
variable "vcenter_password" {}
variable "vcenter_ip" {}
variable "vcenter_dc" {}
variable "vcenter_rp" {}
variable "vcenter_ds" {}

output "internal_cidr" { value = "${var.vsphere_subnet}" }
output "internal_gw" { value = "${var.internal_gw}" }
output "network_name" { value = "${var.network_name}" }
output "vcenter_cluster" { value = "${var.vcenter_cluster}" }
output "jumpbox_url" { value = "${var.jumpbox_ip}:22" }
output "external_ip" { value = "${var.jumpbox_ip}" }
output "jumpbox__internal_ip" { value = "${var.jumpbox_ip}" }
output "director__internal_ip" { value = "${var.bosh_director_internal_ip}" }
output "vcenter_disks" { value = "${var.network_name}" }
output "vcenter_vms" { value = "${var.network_name}_vms" }
output "vcenter_templates" { value = "${var.network_name}_templates" }
output "vcenter_ip" { value = "${var.vcenter_ip}" }
output "vcenter_dc" { value = "${var.vcenter_dc}" }
output "vcenter_rp" { value = "${var.vcenter_rp}" }
output "vcenter_ds" { value = "${var.vcenter_ds}" }
`)
}
